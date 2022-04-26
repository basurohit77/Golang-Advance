// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package acceptor

import "github.com/instana/go-sensor/process"

// ProcessData is a representation of a running process for com.instana.plugin.process plugin
type ProcessData struct {
	PID           int                          `json:"pid"`
	Exec          string                       `json:"exec"`
	Args          []string                     `json:"args,omitempty"`
	Env           map[string]string            `json:"env,omitempty"`
	User          string                       `json:"user,omitempty"`
	Group         string                       `json:"group,omitempty"`
	ContainerID   string                       `json:"container,omitempty"`
	ContainerPid  int                          `json:"containerPid,string,omitempty"`
	ContainerType string                       `json:"containerType,omitempty"`
	Start         int64                        `json:"start"`
	HostName      string                       `json:"com.instana.plugin.host.name"`
	HostPID       int                          `json:"com.instana.plugin.host.pid,string"`
	CPU           *ProcessCPUStatsDelta        `json:"cpu,omitempty"`
	Memory        *ProcessMemoryStatsUpdate    `json:"mem,omitempty"`
	OpenFiles     *ProcessOpenFilesStatsUpdate `json:"openFiles,omitempty"`
}

// NewProcessPluginPayload returns payload for the process plugin of Instana acceptor
func NewProcessPluginPayload(entityID string, data ProcessData) PluginPayload {
	const pluginName = "com.instana.plugin.process"

	return PluginPayload{
		Name:     pluginName,
		EntityID: entityID,
		Data:     data,
	}
}

// ProcessCPUStatsDelta represents the CPU stats that have changed since the last measurement
type ProcessCPUStatsDelta struct {
	User   float64 `json:"user,omitempty"`
	System float64 `json:"sys,omitempty"`
}

// NewProcessCPUStatsDelta calculates the difference between two CPU usage stats.
// It returns nil if stats are equal or if the stats were taken at the same time (ticks).
// The stats are considered to be equal if the change is less then 1%.
func NewProcessCPUStatsDelta(prev, next process.CPUStats, ticksElapsed int) *ProcessCPUStatsDelta {
	if prev == next || ticksElapsed == 0 {
		return nil
	}

	delta := &ProcessCPUStatsDelta{}
	if d := float64(next.System-prev.System) / float64(ticksElapsed); d >= 0.01 {
		delta.System = d
	}
	if d := float64(next.User-prev.User) / float64(ticksElapsed); d >= 0.01 {
		delta.User = d
	}

	if delta.User == 0 && delta.System == 0 {
		return nil
	}

	return delta
}

// ProcessMemoryStatsUpdate represents the memory stats that have changed since the last measurement
type ProcessMemoryStatsUpdate struct {
	Total  *int `json:"virtual,omitempty"`
	Rss    *int `json:"resident,omitempty"`
	Shared *int `json:"share,omitempty"`
}

// NewProcessMemoryStatsUpdate returns the fields that have been updated since the last measurement.
// It returns nil if nothing has changed.
func NewProcessMemoryStatsUpdate(prev, next process.MemStats) *ProcessMemoryStatsUpdate {
	if prev == next {
		return nil
	}

	update := &ProcessMemoryStatsUpdate{}
	if prev.Total != next.Total {
		update.Total = &next.Total
	}
	if prev.Rss != next.Rss {
		update.Rss = &next.Rss
	}
	if prev.Shared != next.Shared {
		update.Shared = &next.Shared
	}

	return update
}

// ProcessOpenFilesStatsUpdate represents the open file stats and limits that have changed since the last measurement
type ProcessOpenFilesStatsUpdate struct {
	Current *int `json:"current,omitempty"`
	Max     *int `json:"max,omitempty"`
}

// NewProcessOpenFilesStatsUpdate returns the (process.ResourceLimits).OpenFiles fields that have been updated
// since the last measurement. It returns nil if nothing has changed.
func NewProcessOpenFilesStatsUpdate(prev, next process.ResourceLimits) *ProcessOpenFilesStatsUpdate {
	if prev.OpenFiles == next.OpenFiles {
		return nil
	}

	update := &ProcessOpenFilesStatsUpdate{}
	if prev.OpenFiles.Current != next.OpenFiles.Current {
		update.Current = &next.OpenFiles.Current
	}
	if prev.OpenFiles.Max != next.OpenFiles.Max {
		update.Max = &next.OpenFiles.Max
	}

	return update
}
