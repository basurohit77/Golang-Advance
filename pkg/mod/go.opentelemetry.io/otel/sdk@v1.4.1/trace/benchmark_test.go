// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func BenchmarkSpanSetAttributesOverCapacity(b *testing.B) {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanLimits(sdktrace.SpanLimits{AttributeCountLimit: 1}),
	)
	tracer := tp.Tracer("BenchmarkSpanSetAttributesOverCapacity")
	ctx := context.Background()
	attrs := make([]attribute.KeyValue, 128)
	for i := range attrs {
		key := fmt.Sprintf("key-%d", i)
		attrs[i] = attribute.Bool(key, true)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, span := tracer.Start(ctx, "/foo")
		span.SetAttributes(attrs...)
		span.End()
	}
}

func BenchmarkStartEndSpan(b *testing.B) {
	traceBenchmark(b, "Benchmark StartEndSpan", func(b *testing.B, t trace.Tracer) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, span := t.Start(ctx, "/foo")
			span.End()
		}
	})
}

func BenchmarkSpanWithAttributes_4(b *testing.B) {
	traceBenchmark(b, "Benchmark Start With 4 Attributes", func(b *testing.B, t trace.Tracer) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := t.Start(ctx, "/foo")
			span.SetAttributes(
				attribute.Bool("key1", false),
				attribute.String("key2", "hello"),
				attribute.Int64("key3", 123),
				attribute.Float64("key4", 123.456),
			)
			span.End()
		}
	})
}

func BenchmarkSpanWithAttributes_8(b *testing.B) {
	traceBenchmark(b, "Benchmark Start With 8 Attributes", func(b *testing.B, t trace.Tracer) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := t.Start(ctx, "/foo")
			span.SetAttributes(
				attribute.Bool("key1", false),
				attribute.String("key2", "hello"),
				attribute.Int64("key3", 123),
				attribute.Float64("key4", 123.456),
				attribute.Bool("key21", false),
				attribute.String("key22", "hello"),
				attribute.Int64("key23", 123),
				attribute.Float64("key24", 123.456),
			)
			span.End()
		}
	})
}

func BenchmarkSpanWithAttributes_all(b *testing.B) {
	traceBenchmark(b, "Benchmark Start With all Attribute types", func(b *testing.B, t trace.Tracer) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := t.Start(ctx, "/foo")
			span.SetAttributes(
				attribute.Bool("key1", false),
				attribute.String("key2", "hello"),
				attribute.Int64("key3", 123),
				attribute.Float64("key7", 123.456),
				attribute.Int("key9", 123),
			)
			span.End()
		}
	})
}

func BenchmarkSpanWithAttributes_all_2x(b *testing.B) {
	traceBenchmark(b, "Benchmark Start With all Attributes types twice", func(b *testing.B, t trace.Tracer) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := t.Start(ctx, "/foo")
			span.SetAttributes(
				attribute.Bool("key1", false),
				attribute.String("key2", "hello"),
				attribute.Int64("key3", 123),
				attribute.Float64("key7", 123.456),
				attribute.Int("key10", 123),
				attribute.Bool("key21", false),
				attribute.String("key22", "hello"),
				attribute.Int64("key23", 123),
				attribute.Float64("key27", 123.456),
				attribute.Int("key210", 123),
			)
			span.End()
		}
	})
}

func BenchmarkSpanWithEvents_4(b *testing.B) {
	traceBenchmark(b, "Benchmark Start With 4 Events", func(b *testing.B, t trace.Tracer) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := t.Start(ctx, "/foo")
			span.AddEvent("event1")
			span.AddEvent("event2")
			span.AddEvent("event3")
			span.AddEvent("event4")
			span.End()
		}
	})
}

func BenchmarkSpanWithEvents_8(b *testing.B) {
	traceBenchmark(b, "Benchmark Start With 4 Events", func(b *testing.B, t trace.Tracer) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := t.Start(ctx, "/foo")
			span.AddEvent("event1")
			span.AddEvent("event2")
			span.AddEvent("event3")
			span.AddEvent("event4")
			span.AddEvent("event5")
			span.AddEvent("event6")
			span.AddEvent("event7")
			span.AddEvent("event8")
			span.End()
		}
	})
}

func BenchmarkSpanWithEvents_WithStackTrace(b *testing.B) {
	traceBenchmark(b, "Benchmark Start With 4 Attributes", func(b *testing.B, t trace.Tracer) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := t.Start(ctx, "/foo")
			span.AddEvent("event1", trace.WithStackTrace(true))
			span.End()
		}
	})
}
func BenchmarkSpanWithEvents_WithTimestamp(b *testing.B) {
	traceBenchmark(b, "Benchmark Start With 4 Attributes", func(b *testing.B, t trace.Tracer) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := t.Start(ctx, "/foo")
			span.AddEvent("event1", trace.WithTimestamp(time.Unix(0, 0)))
			span.End()
		}
	})
}

func BenchmarkTraceID_DotString(b *testing.B) {
	t, _ := trace.TraceIDFromHex("0000000000000001000000000000002a")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: t})

	want := "0000000000000001000000000000002a"
	for i := 0; i < b.N; i++ {
		if got := sc.TraceID().String(); got != want {
			b.Fatalf("got = %q want = %q", got, want)
		}
	}
}

func BenchmarkSpanID_DotString(b *testing.B) {
	sc := trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{1}})
	want := "0100000000000000"
	for i := 0; i < b.N; i++ {
		if got := sc.SpanID().String(); got != want {
			b.Fatalf("got = %q want = %q", got, want)
		}
	}
}

func traceBenchmark(b *testing.B, name string, fn func(*testing.B, trace.Tracer)) {
	b.Run("AlwaysSample", func(b *testing.B) {
		b.ReportAllocs()
		fn(b, tracer(b, name, sdktrace.AlwaysSample()))
	})
	b.Run("NeverSample", func(b *testing.B) {
		b.ReportAllocs()
		fn(b, tracer(b, name, sdktrace.NeverSample()))
	})
}

func tracer(b *testing.B, name string, sampler sdktrace.Sampler) trace.Tracer {
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sampler))
	return tp.Tracer(name)
}
