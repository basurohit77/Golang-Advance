// Copyright 2019, OpenTelemetry Authors
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

package export

import (
	"time"

	"google.golang.org/grpc/codes"

	"go.opentelemetry.io/otel/api/core"
	apitrace "go.opentelemetry.io/otel/api/trace"
)

// SpanData contains all the information collected by a span.
type SpanData struct {
	SpanContext  core.SpanContext
	ParentSpanID core.SpanID
	SpanKind     apitrace.SpanKind
	Name         string
	StartTime    time.Time
	// The wall clock time of EndTime will be adjusted to always be offset
	// from StartTime by the duration of the span.
	EndTime                  time.Time
	Attributes               []core.KeyValue
	MessageEvents            []Event
	Links                    []apitrace.Link
	Status                   codes.Code
	HasRemoteParent          bool
	DroppedAttributeCount    int
	DroppedMessageEventCount int
	DroppedLinkCount         int

	// ChildSpanCount holds the number of child span created for this span.
	ChildSpanCount int
}

// Event is used to describe an Event with a message string and set of
// Attributes.
type Event struct {
	// Message describes the Event.
	Message string

	// Attributes contains a list of keyvalue pairs.
	Attributes []core.KeyValue

	// Time is the time at which this event was recorded.
	Time time.Time
}
