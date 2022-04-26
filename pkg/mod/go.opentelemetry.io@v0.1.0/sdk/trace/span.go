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

package trace

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc/codes"

	"go.opentelemetry.io/otel/api/core"
	apitrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/sdk/export"
	"go.opentelemetry.io/otel/sdk/internal"
)

// span implements apitrace.Span interface.
type span struct {
	// data contains information recorded about the span.
	//
	// It will be non-nil if we are exporting the span or recording events for it.
	// Otherwise, data is nil, and the span is simply a carrier for the
	// SpanContext, so that the trace ID is propagated.
	data        *export.SpanData
	mu          sync.Mutex // protects the contents of *data (but not the pointer value.)
	spanContext core.SpanContext

	// lruAttributes are capped at configured limit. When the capacity is reached an oldest entry
	// is removed to create room for a new entry.
	lruAttributes *lruMap

	// messageEvents are stored in FIFO queue capped by configured limit.
	messageEvents *evictedQueue

	// links are stored in FIFO queue capped by configured limit.
	links *evictedQueue

	// spanStore is the spanStore this span belongs to, if any, otherwise it is nil.
	//*spanStore
	endOnce sync.Once

	executionTracerTaskEnd func()  // ends the execution tracer span
	tracer                 *tracer // tracer used to create span.
}

var _ apitrace.Span = &span{}

func (s *span) SpanContext() core.SpanContext {
	if s == nil {
		return core.EmptySpanContext()
	}
	return s.spanContext
}

func (s *span) IsRecording() bool {
	if s == nil {
		return false
	}
	return s.data != nil
}

func (s *span) SetStatus(status codes.Code) {
	if s == nil {
		return
	}
	if !s.IsRecording() {
		return
	}
	s.mu.Lock()
	s.data.Status = status
	s.mu.Unlock()
}

func (s *span) SetAttribute(attribute core.KeyValue) {
	if !s.IsRecording() {
		return
	}
	s.copyToCappedAttributes(attribute)
}

func (s *span) SetAttributes(attributes ...core.KeyValue) {
	if !s.IsRecording() {
		return
	}
	s.copyToCappedAttributes(attributes...)
}

func (s *span) End(options ...apitrace.EndOption) {
	if s == nil {
		return
	}

	if s.executionTracerTaskEnd != nil {
		s.executionTracerTaskEnd()
	}
	if !s.IsRecording() {
		return
	}
	opts := apitrace.EndOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	s.endOnce.Do(func() {
		sps, _ := s.tracer.provider.spanProcessors.Load().(spanProcessorMap)
		mustExportOrProcess := len(sps) > 0
		if mustExportOrProcess {
			sd := s.makeSpanData()
			if opts.EndTime.IsZero() {
				sd.EndTime = internal.MonotonicEndTime(sd.StartTime)
			} else {
				sd.EndTime = opts.EndTime
			}
			for sp := range sps {
				sp.OnEnd(sd)
			}
		}
	})
}

func (s *span) Tracer() apitrace.Tracer {
	return s.tracer
}

func (s *span) AddEvent(ctx context.Context, msg string, attrs ...core.KeyValue) {
	if !s.IsRecording() {
		return
	}
	s.addEventWithTimestamp(time.Now(), msg, attrs...)
}

func (s *span) AddEventWithTimestamp(ctx context.Context, timestamp time.Time, msg string, attrs ...core.KeyValue) {
	if !s.IsRecording() {
		return
	}
	s.addEventWithTimestamp(timestamp, msg, attrs...)
}

func (s *span) addEventWithTimestamp(timestamp time.Time, msg string, attrs ...core.KeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messageEvents.add(export.Event{
		Message:    msg,
		Attributes: attrs,
		Time:       timestamp,
	})
}

func (s *span) SetName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		// TODO: now what?
		return
	}
	spanName := s.tracer.spanNameWithPrefix(name)
	s.data.Name = spanName
	// SAMPLING
	noParent := !s.data.ParentSpanID.IsValid()
	var ctx core.SpanContext
	if noParent {
		ctx = core.EmptySpanContext()
	} else {
		// FIXME: Where do we get the parent context from?
		// From SpanStore?
		ctx = s.data.SpanContext
	}
	data := samplingData{
		noParent:     noParent,
		remoteParent: s.data.HasRemoteParent,
		parent:       ctx,
		name:         spanName,
		cfg:          s.tracer.provider.config.Load().(*Config),
		span:         s,
	}
	makeSamplingDecision(data)
}

// AddLink implements Span interface. Specified link is added to the span.
// If the total number of links associated with the span exceeds the limit
// then the oldest link is removed to create space for the link being added.
func (s *span) AddLink(link apitrace.Link) {
	if !s.IsRecording() {
		return
	}
	s.addLink(link)
}

// Link implements Span interface. It is similar to AddLink but it excepts
// SpanContext and attributes as arguments instead of Link. It first creates
// a Link object and then adds to the span.
func (s *span) Link(sc core.SpanContext, attrs ...core.KeyValue) {
	if !s.IsRecording() {
		return
	}
	attrsCopy := attrs
	if attrs != nil {
		attrsCopy = make([]core.KeyValue, len(attrs))
		copy(attrsCopy, attrs)
	}
	link := apitrace.Link{SpanContext: sc, Attributes: attrsCopy}
	s.addLink(link)
}

func (s *span) addLink(link apitrace.Link) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links.add(link)
}

// makeSpanData produces a SpanData representing the current state of the span.
// It requires that s.data is non-nil.
func (s *span) makeSpanData() *export.SpanData {
	var sd export.SpanData
	s.mu.Lock()
	defer s.mu.Unlock()
	sd = *s.data
	if s.lruAttributes.simpleLruMap.Len() > 0 {
		sd.Attributes = s.lruAttributesToAttributeMap()
		sd.DroppedAttributeCount = s.lruAttributes.droppedCount
	}
	if len(s.messageEvents.queue) > 0 {
		sd.MessageEvents = s.interfaceArrayToMessageEventArray()
		sd.DroppedMessageEventCount = s.messageEvents.droppedCount
	}
	if len(s.links.queue) > 0 {
		sd.Links = s.interfaceArrayToLinksArray()
		sd.DroppedLinkCount = s.links.droppedCount
	}
	return &sd
}

func (s *span) interfaceArrayToLinksArray() []apitrace.Link {
	linkArr := make([]apitrace.Link, 0)
	for _, value := range s.links.queue {
		linkArr = append(linkArr, value.(apitrace.Link))
	}
	return linkArr
}

func (s *span) interfaceArrayToMessageEventArray() []export.Event {
	messageEventArr := make([]export.Event, 0)
	for _, value := range s.messageEvents.queue {
		messageEventArr = append(messageEventArr, value.(export.Event))
	}
	return messageEventArr
}

func (s *span) lruAttributesToAttributeMap() []core.KeyValue {
	attributes := make([]core.KeyValue, 0, s.lruAttributes.simpleLruMap.Len())
	for _, key := range s.lruAttributes.simpleLruMap.Keys() {
		value, ok := s.lruAttributes.simpleLruMap.Get(key)
		if ok {
			key := key.(core.Key)
			value := value.(core.Value)
			attributes = append(attributes, core.KeyValue{Key: key, Value: value})
		}
	}
	return attributes
}

func (s *span) copyToCappedAttributes(attributes ...core.KeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range attributes {
		s.lruAttributes.add(a.Key, a.Value)
	}
}

func (s *span) addChild() {
	if !s.IsRecording() {
		return
	}
	s.mu.Lock()
	s.data.ChildSpanCount++
	s.mu.Unlock()
}

func startSpanInternal(tr *tracer, name string, parent core.SpanContext, remoteParent bool, o apitrace.SpanOptions) *span {
	var noParent bool
	span := &span{}
	span.spanContext = parent

	cfg := tr.provider.config.Load().(*Config)

	if parent == core.EmptySpanContext() {
		span.spanContext.TraceID = cfg.IDGenerator.NewTraceID()
		noParent = true
	}
	span.spanContext.SpanID = cfg.IDGenerator.NewSpanID()
	data := samplingData{
		noParent:     noParent,
		remoteParent: remoteParent,
		parent:       parent,
		name:         name,
		cfg:          cfg,
		span:         span,
	}
	makeSamplingDecision(data)

	// TODO: [rghetia] restore when spanstore is added.
	// if !internal.LocalSpanStoreEnabled && !span.spanContext.IsSampled() && !o.Record {
	if !span.spanContext.IsSampled() && !o.Record {
		return span
	}

	startTime := o.StartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}
	sk := o.SpanKind
	if sk == "" {
		sk = apitrace.SpanKindInternal
	}
	span.data = &export.SpanData{
		SpanContext:     span.spanContext,
		StartTime:       startTime,
		SpanKind:        sk,
		Name:            name,
		HasRemoteParent: remoteParent,
	}
	span.lruAttributes = newLruMap(cfg.MaxAttributesPerSpan)
	span.messageEvents = newEvictedQueue(cfg.MaxEventsPerSpan)
	span.links = newEvictedQueue(cfg.MaxLinksPerSpan)

	if !noParent {
		span.data.ParentSpanID = parent.SpanID
	}
	// TODO: [rghetia] restore when spanstore is added.
	//if internal.LocalSpanStoreEnabled {
	//	ss := spanStoreForNameCreateIfNew(name)
	//	if ss != nil {
	//		span.spanStore = ss
	//		ss.add(span)
	//	}
	//}

	return span
}

type samplingData struct {
	noParent     bool
	remoteParent bool
	parent       core.SpanContext
	name         string
	cfg          *Config
	span         *span
}

func makeSamplingDecision(data samplingData) {
	if data.noParent || data.remoteParent {
		// If this span is the child of a local span and no
		// Sampler is set in the options, keep the parent's
		// TraceFlags.
		//
		// Otherwise, consult the Sampler in the options if it
		// is non-nil, otherwise the default sampler.
		sampler := data.cfg.DefaultSampler
		//if o.Sampler != nil {
		//	sampler = o.Sampler
		//}
		spanContext := &data.span.spanContext
		sampled := sampler(SamplingParameters{
			ParentContext:   data.parent,
			TraceID:         spanContext.TraceID,
			SpanID:          spanContext.SpanID,
			Name:            data.name,
			HasRemoteParent: data.remoteParent}).Sample
		if sampled {
			spanContext.TraceFlags |= core.TraceFlagsSampled
		} else {
			spanContext.TraceFlags &^= core.TraceFlagsSampled
		}
	}
}
