package ossrecord

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Metric represents one individual source of monitoring data for one service
type Metric struct {
	Type        MetricType `json:"type"`
	PlanOrLabel string     `json:"plan_or_label"`
	Environment string     `json:"environment"`
	Tags        []string   `json:"tags"`
}

// MetricType represents the type of monitoring data reported in one Metric object (e.g. consumption or monitoring)
type MetricType string

// Constants for MetricType
const (
	MetricProvisioning MetricType = "PROVISIONING"
	MetricConsumption  MetricType = "CONSUMPTION"
	MetricOther        MetricType = "OTHERMETRIC"
)

// Constant definitions for common tags associated with Metric objects
// Note: there may be additional, free-form tags
const (
	MetricTagSource    = "Src:"    // How was this Metric found
	MetricTagSourceEDB = "Src:EDB" // This Metric was found in EDB
	MetricTagIssues    = "Issues:" // Number of issues encountered while processing this Metric
)

// AddTags adds one or more tags to a Metric (without duplicates)
func (m *Metric) AddTags(tags ...string) {
	for _, t := range tags {
		var found bool
		for _, t0 := range m.Tags {
			if t0 == t {
				found = true
				break
			}
		}
		if !found {
			m.Tags = append(m.Tags, t)
		}
	}
	sort.Strings(m.Tags)
}

// FindTag checks if a given tag is present in this Metric.
// If the tag exists and ends in a ":", the value after the ":" is returned
// If the tag exists and does not end in a ":", the entire tag string is returned
// If the tag does not exist, the empty string is returned
func (m *Metric) FindTag(tag string) string {
	var pattern *regexp.Regexp

	if strings.HasSuffix(tag, ":") {
		pattern = regexp.MustCompile(fmt.Sprintf(`^%s(\S+)$`, tag))
	}

	for _, t := range m.Tags {
		if pattern != nil {
			if match := pattern.FindStringSubmatch(t); match != nil {
				return match[1]
			}
		} else {
			if t == tag {
				return t
			}
		}
	}
	return ""
}

// String returns a short string representation for this Metric object
func (m *Metric) String() string {
	return (fmt.Sprintf("Metric(%s/%s/%s)", m.Type, m.PlanOrLabel, m.Environment))
}

// ComparableString forces comparisons of Metrics to be on a single line with all attributes together
func (m *Metric) ComparableString() string {
	return m.String()
}
