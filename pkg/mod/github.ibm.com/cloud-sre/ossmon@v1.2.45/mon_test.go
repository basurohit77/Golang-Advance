// IBM Confidential OCO Source Materials
// (C) Copyright and Licensed by IBM Corp. 2017-2021
//
// The source code for this program is not published or otherwise
// divested of its trade secrets, irrespective of what has
// been deposited with the U.S. Copyright Office.

package ossmon

import (
	"context"
	"testing"

	instana "github.com/instana/go-sensor"
	newrelic "github.com/newrelic/go-agent"
	"github.com/opentracing/opentracing-go"
)

var config = newrelic.Config{}
var nrApp newrelic.Application

func init() {
	// config := newrelic.Config{}
	nrApp, _ = newrelic.NewApplication(config)
}

func BenchmarkSetTagsKV(b *testing.B) {
	b.ReportAllocs()
	sensor := instana.NewSensor("bench")
	span := sensor.Tracer().StartSpan("span1")
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		SetTagsKV(ctx, span, "key", "value")
	}
}

// func BenchmarkSetTagsKVWithTransaction(b *testing.B) {
// 	b.ReportAllocs()
// 	sensor := instana.NewSensor("bench")
// 	span := sensor.Tracer().StartSpan("span1")

// 	// ctx := context.Background()
// 	config := newrelic.Config{}
// 	nrApp, err := newrelic.NewApplication(config)
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	txn := nrApp.StartTransaction("txn1", nil, nil)
// 	// var txn newrelic.Transaction

// 	for i := 0; i < b.N; i++ {
// 		SetTagsKVWithTransaction(txn, span, "key", "value")
// 	}
// }

/*
func BenchmarkSetTagsKVWithNone(b *testing.B) {
	sensor := instana.NewSensor("bench")
	span := sensor.Tracer().StartSpan("span1")
	// ctx := context.Background()
	// config := newrelic.Config{}
	// nrApp, err := newrelic.NewApplication(config)
	// if err != nil {
	// 	b.Fatal(err)
	// }
	// txn := nrApp.StartTransaction("txn1", nil, nil)
	// var txn newrelic.Transaction
	for i := 0; i < b.N; i++ {
		SetTagsKVWithNone(span, "key", "value")
	}
}
*/

func TestSetTagsKV(t *testing.T) {
	sensor := instana.NewSensor("bench")

	// type args struct {
	// 	ctx       context.Context
	// 	span      opentracing.Span
	// 	keyValues []interface{}
	// }
	tests := []struct {
		name string
		// args args
		ctx       context.Context
		span      opentracing.Span
		keyValues []interface{}
	}{
		{
			name:      "happy path",
			ctx:       context.Background(),
			span:      sensor.Tracer().StartSpan("span1"),
			keyValues: []interface{}{"k1", "v1", "k2", 2},
		},

		{
			name: "too few tags",
			ctx:  context.Background(),
			span: sensor.Tracer().StartSpan("span1"),
			// keyValues: []string{"k1", "v1"},
			keyValues: []interface{}{"k1"},
		},
		{
			name:      "non string key",
			ctx:       context.Background(),
			span:      sensor.Tracer().StartSpan("span1"),
			keyValues: []interface{}{123, 123},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ctx := WithSpanContext(tt.ctx, tt.span)
			ctx := instana.ContextWithSpan(tt.ctx, tt.span)
			SetTagsKV(ctx, tt.keyValues...)
		})
	}
}

// Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -coverprofile=/var/folders/pf/g_wpbnn9677b6rydzlb_x_v80000gn/T/vscode-goT7lgwU/go-code-cover -bench ^(BenchmarkSetTagsKV|BenchmarkSetTagsKVWithTransaction)$ github.ibm.com/sosat/hooks/mon

// goos: darwin
// goarch: amd64
// pkg: github.ibm.com/sosat/hooks/mon
// BenchmarkSetTagsKV-16                   	15933496	        66.5 ns/op	      16 B/op	       1 allocs/op
// BenchmarkSetTagsKVWithTransaction-16    	13748505	        83.5 ns/op	      16 B/op	       1 allocs/op
// PASS
// coverage: 61.5% of statements
// ok  	github.ibm.com/sosat/hooks/mon	3.083s

// √ hooks %
// √ hooks % go test -bench=. -benchtime 15s -count 2 ./...
// ...
// PASS
// ok  	github.ibm.com/sosat/hooks	1.810s
// goos: darwin
// goarch: amd64
// pkg: github.ibm.com/sosat/hooks/mon
// BenchmarkSetTagsKV-16                   	283994415	        66.6 ns/op
// BenchmarkSetTagsKV-16                   	283210662	        63.1 ns/op
// BenchmarkSetTagsKVWithTransaction-16    	219062989	        80.2 ns/op
// BenchmarkSetTagsKVWithTransaction-16    	221852510	        80.7 ns/op
// PASS
// ok  	github.ibm.com/sosat/hooks/mon	101.763s
// PASS
// ok  	github.ibm.com/sosat/hooks/rate	0.336s
// 2021/03/29 18:02:49 Do failed with error: Post "http://127.0.0.1:61724": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
// 2021/03/29 18:02:49 NewRequest failed with error: parse ":@?@:*": missing protocol scheme
// 2021/03/29 18:02:49 Do failed with error: Post "http://127.0.0.1:61748": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
// 2021/03/29 18:02:49 NewRequest failed with error: parse ":@?@:*": missing protocol scheme
// PASS
// ok  	github.ibm.com/sosat/hooks/vault	0.489s

// another run:
// ...
// PASS
// ok  	github.ibm.com/sosat/hooks	1.062s
// goos: darwin
// goarch: amd64
// pkg: github.ibm.com/sosat/hooks/mon
// BenchmarkSetTagsKV-16                   	302495017	        63.0 ns/op
// BenchmarkSetTagsKV-16                   	292479166	        63.5 ns/op
// BenchmarkSetTagsKVWithTransaction-16    	207590628	        83.1 ns/op
// BenchmarkSetTagsKVWithTransaction-16    	208335974	        85.8 ns/op
// PASS
// ok  	github.ibm.com/sosat/hooks/mon	102.659s
// PASS
// ok  	github.ibm.com/sosat/hooks/rate	0.307s
// 2021/03/29 18:06:07 Do failed with error: Post "http://127.0.0.1:61932": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
// 2021/03/29 18:06:07 NewRequest failed with error: parse ":@?@:*": missing protocol scheme
// 2021/03/29 18:06:07 Do failed with error: Post "http://127.0.0.1:61956": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
// 2021/03/29 18:06:07 NewRequest failed with error: parse ":@?@:*": missing protocol scheme
// PASS
// ok  	github.ibm.com/sosat/hooks/vault	0.153s
// √ hooks %

// BenchmarkSetTagsKV-16                   	266392602	        63.4 ns/op
// BenchmarkSetTagsKV-16                   	281034158	        67.4 ns/op
// BenchmarkSetTagsKVWithTransaction-16    	207067161	        86.5 ns/op
// BenchmarkSetTagsKVWithTransaction-16    	228112242	        78.9 ns/op
// BenchmarkSetTagsKVWithNone-16           	221867167	        82.0 ns/op
// BenchmarkSetTagsKVWithNone-16           	227812627	        80.2 ns/op

// √ mon % go test -bench="BenchmarkSetTagsKVWithNone" -benchtime 15s -count 2 -run=XXX
// goos: darwin
// goarch: amd64
// pkg: github.ibm.com/sosat/hooks/mon
// BenchmarkSetTagsKVWithNone-16    	228724346	        78.3 ns/op
// BenchmarkSetTagsKVWithNone-16    	228531376	        78.7 ns/op
// PASS
// ok  	github.ibm.com/sosat/hooks/mon	52.136s

// √ mon % go test -bench="BenchmarkSetTagsKVWithTransaction" -benchtime 15s -count 2 -run=XXX
// goos: darwin
// goarch: amd64
// pkg: github.ibm.com/sosat/hooks/mon
// BenchmarkSetTagsKVWithTransaction-16    	228623443	        78.7 ns/op
// BenchmarkSetTagsKVWithTransaction-16    	229647218	        78.8 ns/op
// PASS
// ok  	github.ibm.com/sosat/hooks/mon	52.356s
// √ mon %

// √ mon % go test -bench="BenchmarkSetTagsKV" -benchtime 15s -count 2 -run=XXX
// goos: darwin
// goarch: amd64
// pkg: github.ibm.com/sosat/hooks/mon
// BenchmarkSetTagsKV-16                   	295841733	        59.7 ns/op
// BenchmarkSetTagsKV-16                   	300444410	        59.1 ns/op
// BenchmarkSetTagsKVWithTransaction-16

/////

// Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -coverprofile=/var/folders/pf/g_wpbnn9677b6rydzlb_x_v80000gn/T/vscode-goT7lgwU/go-code-cover -bench ^(BenchmarkSetTagsKV|BenchmarkSetTagsKVWithTransaction|BenchmarkSetTagsKVWithNone)$ github.ibm.com/sosat/hooks/mon

// goos: darwin
// goarch: amd64
// pkg: github.ibm.com/sosat/hooks/mon
// BenchmarkSetTagsKV-16                   	18506856	        63.7 ns/op	      16 B/op	       1 allocs/op
// BenchmarkSetTagsKVWithTransaction-16    	13042982	        82.8 ns/op	      16 B/op	       1 allocs/op
// BenchmarkSetTagsKVWithNone-16           	14406897	        81.1 ns/op	      16 B/op	       1 allocs/op
// PASS
// coverage: 47.1% of statements
// ok  	github.ibm.com/sosat/hooks/mon	5.409s

// √ mon % go test -bench="BenchmarkSetTags.*" -benchtime 10s -count 2 -run=XXX
// goos: darwin
// goarch: amd64
// pkg: github.ibm.com/sosat/hooks/mon
// BenchmarkSetTagsKV-16                   	203133165	        60.9 ns/op
// BenchmarkSetTagsKV-16                   	201470432	        58.1 ns/op
// BenchmarkSetTagsKVWithTransaction-16    	151588344	        80.5 ns/op
// BenchmarkSetTagsKVWithTransaction-16    	148732077	        80.8 ns/op
// BenchmarkSetTagsKVWithNone-16           	148274248	        80.7 ns/op
// BenchmarkSetTagsKVWithNone-16           	147169176	        80.5 ns/op
// PASS
// ok  	github.ibm.com/sosat/hooks/mon	117.066s

func TestSetTag(t *testing.T) {
	_, spanctx := NewSpan(context.Background(), instana.NewSensor("test"), "real context")

	// type args struct {
	// 	ctx   context.Context
	// 	key   string
	// 	value interface{}
	// }
	tests := []struct {
		name string
		// args args
		ctx   context.Context
		key   string
		value interface{}
	}{
		{
			name:  "bg context",
			ctx:   context.Background(),
			key:   "k1",
			value: "v1",
		},
		{
			name:  "nil context",
			ctx:   nil,
			key:   "k1",
			value: "v1",
		},
		{
			name: "span in context",
			// ctx:   WithSpanContext(context.Background(), NewSpan(context.Background(), instana.NewSensor("test"), "real context")),
			ctx:   spanctx,
			key:   "k1",
			value: "v1",
		},
		{
			name:  "nil value",
			ctx:   context.Background(),
			key:   "k1",
			value: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetTag(tt.ctx, tt.key, tt.value)
		})
	}
}

func TestSpanFromContext(t *testing.T) {
	_, spanctx := NewSpan(context.Background(), instana.NewSensor("test"), "yes span")
	// type args struct {
	// 	ctx context.Context
	// }
	tests := []struct {
		name string
		// args args
		ctx     context.Context
		wantNil bool
		want    opentracing.Span
	}{
		{
			name:    "no span",
			ctx:     context.Background(),
			wantNil: true,
		},
		{
			name: "yes span",
			// ctx:     WithSpanContext(context.Background(), NewSpan(context.Background(), instana.NewSensor("test"), "yes span")),
			ctx:     spanctx,
			wantNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := SpanFromContext(tt.ctx); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("SpanFromContext() = %v, want %v", got, tt.want)
			// }
			got, _ := instana.SpanFromContext(tt.ctx)
			t.Logf("name=%v\n", tt.name)
			t.Logf("got=%v\n", got)
			t.Logf("wantNil=%v\n", tt.wantNil)

			if got != nil && tt.wantNil == true {
				t.Errorf("SpanFromContext() = %v, wanted nil\n", got)
			}

			if got == nil && !tt.wantNil {
				t.Errorf("SpanFromContext() = %v, did not want nil\n", got)
			}
		})
	}
}
