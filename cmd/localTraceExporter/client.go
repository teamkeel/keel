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

package localTraceExporter

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type client struct {
}

// Compile time check *client implements otlptrace.Client.
var _ otlptrace.Client = (*client)(nil)

// NewClient creates a client that stores the spans in memory.
func NewClient() otlptrace.Client {
	return newClient()
}

func newClient() *client {
	c := &client{}
	return c
}

var traces = make(map[string][]*tracepb.Span)

func GetTrace(traceID string) []*tracepb.Span {
	return traces[traceID]
}

func AllTraces() map[string][]*tracepb.Span {
	return traces
}

func (c *client) Start(ctx context.Context) error {
	return nil
}

func (c *client) Stop(ctx context.Context) error {
	return nil
}

func (c *client) UploadTraces(ctx context.Context, protoSpans []*tracepb.ResourceSpans) error {

	// Unpack all the spans and store in memory by trace ID
	// This is lossy as we're loosing the resource and service data so we may want to improve this later

	for _, ResourceSpan := range protoSpans {
		scopedSpans := ResourceSpan.GetScopeSpans()
		for _, scopedSpans := range scopedSpans {
			for _, span := range scopedSpans.Spans {
				traceID := span.TraceId
				traces[string(traceID)] = append(traces[string(traceID)], span)
			}
		}
	}

	return nil
}

// MarshalLog is the marshaling function used by the logging system to represent this Client.
func (c *client) MarshalLog() interface{} {
	return struct {
		Type string
	}{
		Type: "local",
	}
}
