package util_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/trace"
)

func TestParseTraceparent(t *testing.T) {
	traceparent := "00-71f835dc7ac2750bed2135c7b30dc7fe-b4c9e2a6a0d84702-01"
	spanContext := util.ParseTraceparent(traceparent)

	require.Equal(t, trace.FlagsSampled, spanContext.TraceFlags())
	require.True(t, spanContext.HasTraceID())
	require.True(t, spanContext.HasSpanID())
	require.Equal(t, "71f835dc7ac2750bed2135c7b30dc7fe", spanContext.TraceID().String())
	require.Equal(t, "b4c9e2a6a0d84702", spanContext.SpanID().String())
}

func TestParseTraceparentWithInvalidTraceId(t *testing.T) {
	traceparent := "00-invalid-b4c9e2a6a0d84702-01"
	spanContext := util.ParseTraceparent(traceparent)

	require.False(t, spanContext.IsValid())
	require.Equal(t, trace.SpanContext{}, spanContext)
}

func TestParseTraceparentWithInvalidSpanId(t *testing.T) {
	traceparent := "00-71f835dc7ac2750bed2135c7b30dc7fe-invalid-01"
	spanContext := util.ParseTraceparent(traceparent)

	require.False(t, spanContext.IsValid())
	require.Equal(t, trace.SpanContext{}, spanContext)
}

func TestGetTraceparent(t *testing.T) {
	traceIdBytes, err := hex.DecodeString("71f835dc7ac2750bed2135c7b30dc7fe")
	require.NoError(t, err)
	spanIdBytes, err := hex.DecodeString("b4c9e2a6a0d84702")
	require.NoError(t, err)

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID(traceIdBytes),
		SpanID:     trace.SpanID(spanIdBytes),
		TraceFlags: trace.FlagsSampled,
	})
	require.True(t, spanContext.IsValid())

	traceparent := util.GetTraceparent(spanContext)
	require.Equal(t, "00-71f835dc7ac2750bed2135c7b30dc7fe-b4c9e2a6a0d84702-01", traceparent)
}
