// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package newrelic

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/newrelic/go-agent/v3/internal"
)

func TestResponseCodeIsError(t *testing.T) {
	cfg := config{Config: defaultConfig()}
	cfg.ErrorCollector.IgnoreStatusCodes = append(cfg.ErrorCollector.IgnoreStatusCodes, 504)
	run := newAppRun(cfg, internal.ConnectReplyDefaults())

	for _, tc := range []struct {
		Code    int
		IsError bool
	}{
		{Code: 0, IsError: false}, // gRPC
		{Code: 1, IsError: true},  // gRPC
		{Code: 5, IsError: false}, // gRPC
		{Code: 6, IsError: true},  // gRPC
		{Code: 99, IsError: true},
		{Code: 100, IsError: false},
		{Code: 199, IsError: false},
		{Code: 200, IsError: false},
		{Code: 300, IsError: false},
		{Code: 399, IsError: false},
		{Code: 400, IsError: true},
		{Code: 404, IsError: false},
		{Code: 503, IsError: true},
		{Code: 504, IsError: false},
	} {
		if is := run.responseCodeIsError(tc.Code); is != tc.IsError {
			t.Errorf("responseCodeIsError for %d, wanted=%v got=%v",
				tc.Code, tc.IsError, is)
		}
	}

}

func TestCrossAppTracingEnabled(t *testing.T) {
	// CAT should be enabled by default.
	cfg := config{Config: defaultConfig()}
	run := newAppRun(cfg, internal.ConnectReplyDefaults())
	if enabled := run.Config.CrossApplicationTracer.Enabled; !enabled {
		t.Error(enabled)
	}

	// DT gets priority over CAT.
	cfg = config{Config: defaultConfig()}
	cfg.DistributedTracer.Enabled = true
	cfg.CrossApplicationTracer.Enabled = true
	run = newAppRun(cfg, internal.ConnectReplyDefaults())
	if enabled := run.Config.CrossApplicationTracer.Enabled; enabled {
		t.Error(enabled)
	}

	cfg = config{Config: defaultConfig()}
	cfg.DistributedTracer.Enabled = false
	cfg.CrossApplicationTracer.Enabled = false
	run = newAppRun(cfg, internal.ConnectReplyDefaults())
	if enabled := run.Config.CrossApplicationTracer.Enabled; enabled {
		t.Error(enabled)
	}

	cfg = config{Config: defaultConfig()}
	cfg.DistributedTracer.Enabled = false
	cfg.CrossApplicationTracer.Enabled = true
	run = newAppRun(cfg, internal.ConnectReplyDefaults())
	if enabled := run.Config.CrossApplicationTracer.Enabled; !enabled {
		t.Error(enabled)
	}
}

func TestTxnTraceThreshold(t *testing.T) {
	// Test that the default txn trace threshold is the failing apdex.
	cfg := config{Config: defaultConfig()}
	run := newAppRun(cfg, internal.ConnectReplyDefaults())
	threshold := run.txnTraceThreshold(1 * time.Second)
	if threshold != 4*time.Second {
		t.Error(threshold)
	}

	// Test that the trace threshold can be assigned to a fixed value.
	cfg = config{Config: defaultConfig()}
	cfg.TransactionTracer.Threshold.IsApdexFailing = false
	cfg.TransactionTracer.Threshold.Duration = 3 * time.Second
	run = newAppRun(cfg, internal.ConnectReplyDefaults())
	threshold = run.txnTraceThreshold(1 * time.Second)
	if threshold != 3*time.Second {
		t.Error(threshold)
	}

	// Test that the trace threshold can be overwritten by server-side-config.
	// with "apdex_f".
	cfg = config{Config: defaultConfig()}
	cfg.TransactionTracer.Threshold.IsApdexFailing = false
	cfg.TransactionTracer.Threshold.Duration = 3 * time.Second
	reply := internal.ConnectReplyDefaults()
	json.Unmarshal([]byte(`{"agent_config":{"transaction_tracer.transaction_threshold":"apdex_f"}}`), &reply)
	run = newAppRun(cfg, reply)
	threshold = run.txnTraceThreshold(1 * time.Second)
	if threshold != 4*time.Second {
		t.Error(threshold)
	}

	// Test that the trace threshold can be overwritten by server-side-config.
	// with a numberic value.
	cfg = config{Config: defaultConfig()}
	reply = internal.ConnectReplyDefaults()
	json.Unmarshal([]byte(`{"agent_config":{"transaction_tracer.transaction_threshold":3}}`), &reply)
	run = newAppRun(cfg, reply)
	threshold = run.txnTraceThreshold(1 * time.Second)
	if threshold != 3*time.Second {
		t.Error(threshold)
	}
}

func TestEmptyReplyEventHarvestDefaults(t *testing.T) {
	run := newAppRun(config{Config: defaultConfig()}, &internal.ConnectReply{})
	assertHarvestConfig(t, run.harvestConfig, expectHarvestConfig{
		maxTxnEvents:    internal.MaxTxnEvents,
		maxCustomEvents: internal.MaxCustomEvents,
		maxErrorEvents:  internal.MaxErrorEvents,
		maxSpanEvents:   maxSpanEvents,
		periods: map[harvestTypes]time.Duration{
			harvestTypesAll: 60 * time.Second,
			0:               60 * time.Second,
		},
	})
}

func TestEventHarvestFieldsAllPopulated(t *testing.T) {
	reply, err := internal.UnmarshalConnectReply([]byte(`{"return_value":{
			"event_harvest_config": {
				"report_period_ms": 5000,
				"harvest_limits": {
					"analytic_event_data": 1,
					"custom_event_data": 2,
					"span_event_data": 3,
					"error_event_data": 4
				}
			}
		}}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	run := newAppRun(config{Config: defaultConfig()}, reply)
	assertHarvestConfig(t, run.harvestConfig, expectHarvestConfig{
		maxTxnEvents:    1,
		maxCustomEvents: 2,
		maxErrorEvents:  4,
		maxSpanEvents:   3,
		periods: map[harvestTypes]time.Duration{
			harvestMetricsTraces: 60 * time.Second,
			harvestTypesEvents:   5 * time.Second,
		},
	})
}

func TestZeroReportPeriod(t *testing.T) {
	reply, err := internal.UnmarshalConnectReply([]byte(`{"return_value":{
			"event_harvest_config": {
				"report_period_ms": 0
			}
		}}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	run := newAppRun(config{Config: defaultConfig()}, reply)
	assertHarvestConfig(t, run.harvestConfig, expectHarvestConfig{
		maxTxnEvents:    internal.MaxTxnEvents,
		maxCustomEvents: internal.MaxCustomEvents,
		maxErrorEvents:  internal.MaxErrorEvents,
		maxSpanEvents:   maxSpanEvents,
		periods: map[harvestTypes]time.Duration{
			harvestTypesAll: 60 * time.Second,
			0:               60 * time.Second,
		},
	})
}

func TestEventHarvestFieldsOnlySpanEvents(t *testing.T) {
	reply, err := internal.UnmarshalConnectReply([]byte(`{"return_value":{
			"event_harvest_config": {
				"report_period_ms": 5000,
				"harvest_limits": { "span_event_data": 3 }
			}}}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	run := newAppRun(config{Config: defaultConfig()}, reply)
	assertHarvestConfig(t, run.harvestConfig, expectHarvestConfig{
		maxTxnEvents:    internal.MaxTxnEvents,
		maxCustomEvents: internal.MaxCustomEvents,
		maxErrorEvents:  internal.MaxErrorEvents,
		maxSpanEvents:   3,
		periods: map[harvestTypes]time.Duration{
			harvestTypesAll ^ harvestSpanEvents: 60 * time.Second,
			harvestSpanEvents:                   5 * time.Second,
		},
	})
}

func TestEventHarvestFieldsOnlyTxnEvents(t *testing.T) {
	reply, err := internal.UnmarshalConnectReply([]byte(`{"return_value":{
			"event_harvest_config": {
				"report_period_ms": 5000,
				"harvest_limits": { "analytic_event_data": 3 }
			}}}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	run := newAppRun(config{Config: defaultConfig()}, reply)
	assertHarvestConfig(t, run.harvestConfig, expectHarvestConfig{
		maxTxnEvents:    3,
		maxCustomEvents: internal.MaxCustomEvents,
		maxErrorEvents:  internal.MaxErrorEvents,
		maxSpanEvents:   maxSpanEvents,
		periods: map[harvestTypes]time.Duration{
			harvestTypesAll ^ harvestTxnEvents: 60 * time.Second,
			harvestTxnEvents:                   5 * time.Second,
		},
	})
}

func TestEventHarvestFieldsOnlyErrorEvents(t *testing.T) {
	reply, err := internal.UnmarshalConnectReply([]byte(`{"return_value":{
			"event_harvest_config": {
				"report_period_ms": 5000,
				"harvest_limits": { "error_event_data": 3 }
			}}}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	run := newAppRun(config{Config: defaultConfig()}, reply)
	assertHarvestConfig(t, run.harvestConfig, expectHarvestConfig{
		maxTxnEvents:    internal.MaxTxnEvents,
		maxCustomEvents: internal.MaxCustomEvents,
		maxErrorEvents:  3,
		maxSpanEvents:   maxSpanEvents,
		periods: map[harvestTypes]time.Duration{
			harvestTypesAll ^ harvestErrorEvents: 60 * time.Second,
			harvestErrorEvents:                   5 * time.Second,
		},
	})
}

func TestEventHarvestFieldsOnlyCustomEvents(t *testing.T) {
	reply, err := internal.UnmarshalConnectReply([]byte(`{"return_value":{
			"event_harvest_config": {
				"report_period_ms": 5000,
				"harvest_limits": { "custom_event_data": 3 }
			}}}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	run := newAppRun(config{Config: defaultConfig()}, reply)
	assertHarvestConfig(t, run.harvestConfig, expectHarvestConfig{
		maxTxnEvents:    internal.MaxTxnEvents,
		maxCustomEvents: 3,
		maxErrorEvents:  internal.MaxErrorEvents,
		maxSpanEvents:   maxSpanEvents,
		periods: map[harvestTypes]time.Duration{
			harvestTypesAll ^ harvestCustomEvents: 60 * time.Second,
			harvestCustomEvents:                   5 * time.Second,
		},
	})
}

func TestConfigurableHarvestNegativeReportPeriod(t *testing.T) {
	h, err := internal.UnmarshalConnectReply([]byte(`{"return_value":{
			"event_harvest_config": {
				"report_period_ms": -1
			}}}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	expect := time.Duration(internal.DefaultConfigurableEventHarvestMs) * time.Millisecond
	if period := h.ConfigurablePeriod(); period != expect {
		t.Fatal(expect, period)
	}
}

func TestReplyTraceIDGenerator(t *testing.T) {
	// Test that the default connect reply has a populated trace id
	// generator that works.
	reply := internal.ConnectReplyDefaults()
	id1 := reply.TraceIDGenerator.GenerateTraceID()
	id2 := reply.TraceIDGenerator.GenerateTraceID()
	if len(id1) != 32 || len(id2) != 32 || id1 == id2 {
		t.Error(id1, id2)
	}
	spanID1 := reply.TraceIDGenerator.GenerateSpanID()
	spanID2 := reply.TraceIDGenerator.GenerateSpanID()
	if len(spanID1) != 16 || len(spanID2) != 16 || spanID1 == spanID2 {
		t.Error(spanID1, spanID2)
	}
}

func TestConfigurableTxnEvents_withCollResponse(t *testing.T) {
	h, err := internal.UnmarshalConnectReply([]byte(
		`{"return_value":{
			"event_harvest_config": {
				"report_period_ms": 10000,
                "harvest_limits": {
             		"analytic_event_data": 15
                }
			}
        }}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	result := newAppRun(config{Config: defaultConfig()}, h).MaxTxnEvents()
	if result != 15 {
		t.Error(fmt.Sprintf("Unexpected max number of txn events, expected %d but got %d", 15, result))
	}
}

func TestConfigurableTxnEvents_notInCollResponse(t *testing.T) {
	reply, err := internal.UnmarshalConnectReply([]byte(
		`{"return_value":{
			"event_harvest_config": {
				"report_period_ms": 10000
			}
        }}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	expected := 10
	cfg := config{Config: defaultConfig()}
	cfg.TransactionEvents.MaxSamplesStored = expected
	result := newAppRun(cfg, reply).MaxTxnEvents()
	if result != expected {
		t.Error(fmt.Sprintf("Unexpected max number of txn events, expected %d but got %d", expected, result))
	}
}

func TestConfigurableTxnEvents_configMoreThanMax(t *testing.T) {
	h, err := internal.UnmarshalConnectReply([]byte(
		`{"return_value":{
			"event_harvest_config": {
				"report_period_ms": 10000
			}
        }}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	cfg := config{Config: defaultConfig()}
	cfg.TransactionEvents.MaxSamplesStored = internal.MaxTxnEvents + 100
	result := newAppRun(cfg, h).MaxTxnEvents()
	if result != internal.MaxTxnEvents {
		t.Error(fmt.Sprintf("Unexpected max number of txn events, expected %d but got %d", internal.MaxTxnEvents, result))
	}
}

type expectHarvestConfig struct {
	maxTxnEvents    int
	maxCustomEvents int
	maxErrorEvents  int
	maxSpanEvents   int
	periods         map[harvestTypes]time.Duration
}

func assertHarvestConfig(t testing.TB, hc harvestConfig, expect expectHarvestConfig) {
	if h, ok := t.(interface {
		Helper()
	}); ok {
		h.Helper()
	}
	if max := hc.MaxTxnEvents; max != expect.maxTxnEvents {
		t.Error(max, expect.maxTxnEvents)
	}
	if max := hc.MaxCustomEvents; max != expect.maxCustomEvents {
		t.Error(max, expect.maxCustomEvents)
	}
	if max := hc.MaxSpanEvents; max != expect.maxSpanEvents {
		t.Error(max, expect.maxSpanEvents)
	}
	if max := hc.MaxErrorEvents; max != expect.maxErrorEvents {
		t.Error(max, expect.maxErrorEvents)
	}
	if periods := hc.ReportPeriods; !reflect.DeepEqual(periods, expect.periods) {
		t.Error(periods, expect.periods)
	}
}

func TestPlaceholderAppRunSampler(t *testing.T) {
	// Test that the placeholder run used before connect does not sample
	// transactions.
	run := newPlaceholderAppRun(config{Config: defaultConfig()})
	if sampled := run.adaptiveSampler.computeSampled(1.0, time.Now()); sampled {
		t.Fatal(sampled)
	}
}

func TestAppRunSampler(t *testing.T) {
	// Test that a default app run samples transactions.
	// Test that the default txn trace threshold is the failing apdex.
	cfg := config{Config: defaultConfig()}
	run := newAppRun(cfg, internal.ConnectReplyDefaults())
	if sampled := run.adaptiveSampler.computeSampled(1.0, time.Now()); !sampled {
		t.Fatal(sampled)
	}
	if run.adaptiveSampler.target != 10 || run.adaptiveSampler.period != 60*time.Second {
		t.Fatal("invalid sampler initialization",
			run.adaptiveSampler.target, run.adaptiveSampler.period)
	}
}

func TestCreateTransactionName(t *testing.T) {
	reply, err := internal.UnmarshalConnectReply([]byte(`{"return_value":{
		"url_rules":[
			{"match_expression":"zip","each_segment":true,"replacement":"zoop"}
		],
		"transaction_name_rules":[
			{"match_expression":"WebTransaction/Go/zap/zoop/zep",
			 "replacement":"WebTransaction/Go/zap/zoop/zep/zup/zyp"}
		],
		"transaction_segment_terms":[
			{"prefix": "WebTransaction/Go/",
			 "terms": ["zyp", "zoop", "zap"]}
		]
	}}`), internal.PreconnectReply{})
	if nil != err {
		t.Fatal(err)
	}
	run := newAppRun(config{Config: defaultConfig()}, reply)

	want := "WebTransaction/Go/zap/zoop/*/zyp"
	if out := run.createTransactionName("/zap/zip/zep", true); out != want {
		t.Error("wanted:", want, "got:", out)
	}
	// Check that the cache was populated as expected.
	if out := run.rulesCache.find("/zap/zip/zep", true); out != want {
		t.Error("wanted:", want, "got:", out)
	}
	// Check that the next call returns the same output.
	if out := run.createTransactionName("/zap/zip/zep", true); out != want {
		t.Error("wanted:", want, "got:", out)
	}
}
