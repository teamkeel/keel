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
	"encoding/hex"
	"maps"
	"sync"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type client struct{}

// Compile time check *client implements otlptrace.Client.
var _ otlptrace.Client = (*client)(nil)

var (
	mu           sync.Mutex
	traces       = make(map[string][]*tracepb.Span)
	traceSummary = make(map[string]*TraceSummary)
)

type TraceSummary struct {
	StartTime time.Time
	EndTime   time.Time
	HasError  bool
	Duration  time.Duration
	RootName  string
	Type      string
}

// NewClient creates a client that stores the spans in memory.
func NewClient() otlptrace.Client {
	return newClient()
}

func newClient() *client {
	c := &client{}
	return c
}

func GetTrace(traceID string) []*tracepb.Span {
	mu.Lock()
	defer mu.Unlock()

	return traces[traceID]
}

func Summary() map[string]*TraceSummary {
	mu.Lock()
	defer mu.Unlock()

	cpy := map[string]*TraceSummary{}
	maps.Copy(cpy, traceSummary)

	return cpy
}

func (c *client) Start(ctx context.Context) error {
	return nil
}

func (c *client) Stop(ctx context.Context) error {
	return nil
}

func (c *client) UploadTraces(ctx context.Context, protoSpans []*tracepb.ResourceSpans) error {
	mu.Lock()
	defer mu.Unlock()

	// Unpack all the spans and store in memory by trace ID
	// This is lossy as we're loosing the resource and service data so we may want to improve this later
	for _, ResourceSpan := range protoSpans {
		scopedSpans := ResourceSpan.GetScopeSpans()
		for _, scopedSpans := range scopedSpans {
			for _, span := range scopedSpans.GetSpans() {
				traceID := hex.EncodeToString(span.GetTraceId())
				traces[traceID] = append(traces[traceID], span)

				start := time.Unix(0, int64(span.GetStartTimeUnixNano()))
				end := time.Unix(0, int64(span.GetEndTimeUnixNano()))

				summary, has := traceSummary[traceID]
				if !has {
					summary = &TraceSummary{
						StartTime: start,
						EndTime:   end,
						Duration:  end.Sub(start),
						HasError:  false,
					}
				} else {
					if start.Before(summary.StartTime) {
						summary.StartTime = start
					}
					if end.After(summary.EndTime) {
						summary.EndTime = end
					}

					summary.Duration = summary.EndTime.Sub(summary.StartTime)
				}

				if span.ParentSpanId == nil {
					summary.RootName = span.GetName()

					for _, attr := range span.GetAttributes() {
						if attr.GetKey() == "type" {
							summary.Type = attr.GetValue().GetStringValue()
						}
					}
				}

				if span.GetStatus().GetCode() == tracepb.Status_STATUS_CODE_ERROR {
					summary.HasError = true
				}

				traceSummary[traceID] = summary
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
