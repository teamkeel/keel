package util

import (
	"encoding/hex"
	"fmt"
	"regexp"

	"go.opentelemetry.io/otel/trace"
)

var traceCtxRegExp = regexp.MustCompile("^(?P<version>[0-9a-f]{2})-(?P<traceID>[a-f0-9]{32})-(?P<spanID>[a-f0-9]{16})-(?P<traceFlags>[a-f0-9]{2})(?:-.*)?$")

// Retrieves traceparent, or empty string if no span exists.
// Uses standard format: {version}-{trace_id}-{span_id}-{trace_flags}
// See specification: https://www.w3.org/TR/trace-context/#traceparent-header-field-values
func GetTraceparent(spanContext trace.SpanContext) string {
	if !spanContext.IsValid() {
		return ""
	}

	// Clear all flags other than the trace-context supported sampling bit.
	flags := spanContext.TraceFlags() & trace.FlagsSampled

	return fmt.Sprintf("%s-%s-%s-%s",
		"00",
		spanContext.TraceID(),
		spanContext.SpanID(),
		flags)
}

func ParseTraceparent(traceparent string) trace.SpanContext {
	maxVersion := 254

	if traceparent == "" {
		return trace.SpanContext{}
	}

	if traceparent == "" {
		return trace.SpanContext{}
	}

	matches := traceCtxRegExp.FindStringSubmatch(traceparent)

	if len(matches) == 0 {
		return trace.SpanContext{}
	}

	if len(matches) < 5 { // four subgroups plus the overall match
		return trace.SpanContext{}
	}

	if len(matches[1]) != 2 {
		return trace.SpanContext{}
	}
	ver, err := hex.DecodeString(matches[1])
	if err != nil {
		return trace.SpanContext{}
	}
	version := int(ver[0])
	if version > maxVersion {
		return trace.SpanContext{}
	}

	if version == 0 && len(matches) != 5 { // four subgroups plus the overall match
		return trace.SpanContext{}
	}

	if len(matches[2]) != 32 {
		return trace.SpanContext{}
	}

	var scc trace.SpanContextConfig

	scc.TraceID, err = trace.TraceIDFromHex(matches[2][:32])
	if err != nil {
		return trace.SpanContext{}
	}

	if len(matches[3]) != 16 {
		return trace.SpanContext{}
	}
	scc.SpanID, err = trace.SpanIDFromHex(matches[3])
	if err != nil {
		return trace.SpanContext{}
	}

	if len(matches[4]) != 2 {
		return trace.SpanContext{}
	}
	opts, err := hex.DecodeString(matches[4])
	if err != nil || len(opts) < 1 || (version == 0 && opts[0] > 2) {
		return trace.SpanContext{}
	}

	// Clear all flags other than the trace-context supported sampling bit.
	scc.TraceFlags = trace.TraceFlags(opts[0]) & trace.FlagsSampled
	scc.Remote = true

	sc := trace.NewSpanContext(scc)
	if !sc.IsValid() {
		return trace.SpanContext{}
	}

	return sc
}
