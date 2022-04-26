package stats

import (
	"fmt"
)

// Counter holds a set of statistics, initialized by `NewStatsCounter()`.
// Channels are used to synchronize statistics read and writes.
// Writes should wait for the `triggerStatsChan` to confirm that the write has happened
// since this helps in situations where a write happens and then immediately a read happens.
// Reads write to `triggerGetChan` and wait for the worker to return stats on `getStatsChan`
type Counter struct {
	counters  map[string]int // map of statistics counters
	readChan  chan readStat  // channel used for receiving statistics
	writeChan chan writeStat // channel used for sending statistics
	//resetChan chan bool	// saving these since they are being considered for the future
	//startTime      time.Time
}

type readStat struct {
	stats chan map[string]int
}

type writeStat struct {
	stat         stat
	confirmWrite chan struct{}
}

type stat struct {
	name  string
	value int
}

// NewStatsCounter returns a new StatsCounter
func NewStatsCounter() *Counter {
	var s Counter
	s.counters = map[string]int{}
	s.readChan = make(chan readStat)
	s.writeChan = make(chan writeStat)

	go s.statWorker()

	return &s
}

// Inc will increase the specified statistic `stat` by 1
// if the `stat` does not exist, it will be registered
func (c *Counter) Inc(statName string) {
	c.IncByValue(statName, 1)
}

// IncByValue will increase the specified statistic `stat` by the specified `value` integer
// value can be negative
// if the `stat` does not exist, it will be registered and set
func (c *Counter) IncByValue(statName string, value int) {
	write := writeStat{
		stat:         stat{name: statName, value: value},
		confirmWrite: make(chan struct{}),
	}
	c.writeChan <- write
	<-write.confirmWrite
}

// ReportStats returns a merged JSON object of the provided statistics with their respective values
// if multiple statistics maps are provided and duplicates are found the function will return an error
// func ReportStats(statsToReport ...map[string]int) ([]byte, error) {
// 	str, err := json.Marshal(counters)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return str, nil
// }

// GetStat returns the value of the statistic with the specified name
// returns zero and an error if the stat with the provided name is not found
func (c *Counter) GetStat(stat string) (int, error) {
	statsMap := c.GetStatsAsMap()
	if statToReturn, ok := statsMap[stat]; ok {
		return statToReturn, nil
	}

	return 0, fmt.Errorf("error retrieving stat: %s not found", stat)
}

// GetStatsAsMap returns all of the recorder statistics in the form of a map[string]int
func (c *Counter) GetStatsAsMap() map[string]int {
	read := readStat{
		stats: make(chan map[string]int),
	}
	c.readChan <- read
	return <-read.stats
}

// single place where the global counters are accessed, values are passed through
// channels to setter and getter
func (c *Counter) statWorker() {
	for {
		select {
		case write := <-c.writeChan:
			c.counters[write.stat.name] += write.stat.value
			write.confirmWrite <- struct{}{}
		case read := <-c.readChan:
			counters := map[string]int{}
			for k, v := range c.counters {
				counters[k] = v
			}
			read.stats <- counters
		}
	}
}
