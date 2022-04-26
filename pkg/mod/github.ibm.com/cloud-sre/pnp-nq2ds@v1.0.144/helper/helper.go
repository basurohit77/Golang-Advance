package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"time"

	"github.com/araddon/dateparse"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

var (
	Debug = false
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func InArrayObj(val interface{}, array interface{}) (exists bool) {
	exists = false
	//index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				//index = i
				exists = true
				return
			}
		}
	}
	return
}

// Removes the WatchReturn from the array by swapping out oisitions and shortening
func RemoveWatchFromArray(s []datastore.WatchReturn, i int) []datastore.WatchReturn {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

// Checks whether the record ID is in the WatchReturn array
func RecordInWatchArray(val string, watchArray []datastore.WatchReturn) (exists bool, index int) {
	exists = false
	index = -1
	for i := 0; i < len(watchArray); i++ {
		if val == watchArray[i].RecordID {
			index = i
			exists = true
			return
		}
	}

	return
}
func RemoveSubscriptionReturnFromArray(s []datastore.SubscriptionReturn, i int) []datastore.SubscriptionReturn {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

// Checks whether the subscription is in the WatchReturn array
func SubscriptionInWatchArray(val string, watchArray []datastore.WatchReturn) (exists bool) {
	exists = false
	for i := 0; i < len(watchArray); i++ {
		if val == watchArray[i].SubscriptionURL.URL {
			//index = i
			exists = true
			return
		}
	}

	return
}
func Restarter(panics int, id string, f func()) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(tlog.Log()+"Not able to recover :", id)
			fmt.Println(tlog.Log(), err)
			if panics > 10 {
				panic("Too many panics.  Do something.")
			} else {
				time.Sleep(time.Duration(panics) * time.Second)
				go Restarter(panics+1, id, f)
			}
		}
	}()
	f()
}

// GetJson returns the value in json
func GetJson(Object interface{}) string {

	bytes, err := json.Marshal(Object)
	FailOnError(err, tlog.Log()+"Failure marshaling ")
	return string(bytes)
}

// GetPrettyJson returns the value in json
func GetPrettyJson(Object interface{}) string {

	bytes, err := json.MarshalIndent(Object, "", "	")
	FailOnError(err, tlog.Log()+"Failure marshaling ")
	return string(bytes)
}
func UnmarshalAnything(instr string) interface{} {

	b := []byte(instr)
	var f interface{}
	err := json.Unmarshal(b, &f)
	FailOnError(err, tlog.Log()+"Failure Unmarshaling ")
	return f
}

// IsNewTimeBeforeExistingTime - checks if the provided newTime is before the existing time.
// Uses the layouts provided to parse the times.
func IsNewTimeBeforeExistingTime(newTime string, newTimeLayout string, existingTime string, existingTimeLayout string) bool {

	//log.Println(tlog.Log(), "newtime: ", newTime, "existing time:", existingTime)
	c, _ := CompareTime(newTime, newTimeLayout, existingTime, existingTimeLayout)
	return c < 0
}

// IsNewTimeAfterExistingTime - checks if the provided newTime is after the existing time.
// Uses the layouts provided to parse the times.
func IsNewTimeAfterExistingTime(newTime string, newTimeLayout string, existingTime string, existingTimeLayout string) bool {

	//log.Println(tlog.Log(), "newtime: ", newTime, "existing time:", existingTime)
	c, _ := CompareTime(newTime, newTimeLayout, existingTime, existingTimeLayout)
	return c > 0
}

// CompareTime - Compares two times for equality.
// result == 0 : times are the same
//
// result > 0 : new time is after existing time
//
// result < 0 : new time is before existing time
//
// Uses the layouts provided to parse the times.
// Error returned if both times are in bad format.
func CompareTime(newTime string, newTimeLayout string, existingTime string, existingTimeLayout string) (int, error) {

	const tformat = "2006-01-02 15:04:05Z07"

	var dnew time.Time
	var dexisting time.Time
	var existingTimeParseErr, newTimeParseErr error

	loc, err := time.LoadLocation("")
	if err != nil {
		panic(err.Error())
	}
	time.Local = loc

	var re = regexp.MustCompile(`(?m)(\d{4}-[01]\d-[0-3]\d) ([0-2]\d:[0-5]\d:[0-5]\d)(?:\.\d+)?([+-]?[01]\d)(:?\d\d)?`)
	if re.Match([]byte(newTime)) {
		subMatches := re.FindStringSubmatch(newTime)
		newTime = fmt.Sprint(subMatches[1], " ", subMatches[2], subMatches[3])
		newTimeLayout = tformat
		dnew, newTimeParseErr = time.Parse(newTimeLayout, newTime)
	} else if newTimeLayout != "" {
		dnew, newTimeParseErr = time.Parse(newTimeLayout, newTime)
	} else {
		dnew, newTimeParseErr = dateparse.ParseLocal(newTime)
	}
	if newTimeParseErr != nil {
		log.Print("Error parsing newTime: ", newTime, "\n", newTimeParseErr)
	}

	if re.Match([]byte(existingTime)) {
		existingTimeLayout = tformat
		subMatches := re.FindStringSubmatch(existingTime)
		existingTime = fmt.Sprint(subMatches[1], " ", subMatches[2], subMatches[3])
		dexisting, existingTimeParseErr = time.Parse(existingTimeLayout, existingTime)
	} else if existingTimeLayout != "" {
		dexisting, existingTimeParseErr = time.Parse(existingTimeLayout, existingTime)
	} else {
		dexisting, existingTimeParseErr = dateparse.ParseLocal(existingTime)
	}
	if existingTimeParseErr != nil {
		log.Print("Error parsing existingTime: ", existingTime, "\n", existingTimeParseErr)
	}

	if existingTimeParseErr != nil && newTimeParseErr == nil {
		// new time is OK, but existing time is not
		return 1, nil
	} else if existingTimeParseErr == nil && newTimeParseErr != nil {
		// existing time is OK, but new time is not
		return -1, nil
	} else if existingTimeParseErr != nil && newTimeParseErr != nil {
		// both times are not OK, if both times are bad, return error
		return 0, errors.New("two bad time formats")
	}

	dnewMs := dnew.UnixNano() / 1000000
	dexistMs := dexisting.UnixNano() / 1000000

	log.Println(tlog.Log()+"DEBUG:", dnewMs, dexistMs)
	log.Println(tlog.Log()+"DEBUG:", dnew, dexisting)
	isNewer := dnew.After(dexisting)
	isOlder := dnew.Before(dexisting)
	isSame := dnew.Equal(dexisting)
	log.Println("DEBUG:", tlog.Log(), "\nNew:", dnew, "\nExisting:", dexisting, "\nnewer?:", isNewer)

	if isSame {
		return 0, nil
	} else if isOlder {
		return -1, nil
	}

	return 1, nil // New time must be after
}

func Dbg(out ...interface{}) {
	if Debug {
		log.Println(out...)
	}
}
