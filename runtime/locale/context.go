package locale

import (
	"context"
	"fmt"
	"net/http"
	"time"
	_ "time/tzdata"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	locationContextKey contextKey = "location"
)

// WithTimeLocation sets the given time location on the context
func WithTimeLocation(ctx context.Context, location time.Location) context.Context {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("timezone.location", location.String()))

	return context.WithValue(ctx, locationContextKey, location)
}

// GetTimeLocation returns the time location set on the context, or error if none set
func GetTimeLocation(ctx context.Context) (time.Location, error) {
	v, ok := ctx.Value(locationContextKey).(time.Location)
	if !ok {
		return time.Location{}, fmt.Errorf("context does not have a key or is not a time.Location: %s", locationContextKey)
	}
	return v, nil
}

// HasTimeLocation returns true if the context has a Time Location set
func HasTimeLocation(ctx context.Context) bool {
	_, err := GetTimeLocation(ctx)

	return err == nil
}

// HandleTimezoneHeader will extract the time location from the Time-Zone header and set it on the context.
// The location will also be set as an attribute to the context's tracing span.
// If no header set, the location will default to UTC
func HandleTimezoneHeader(ctx context.Context, headers http.Header) (time.Location, error) {
	// if no header, a UTC location will be returned
	location, err := time.LoadLocation(headers.Get("Time-Zone"))
	if err != nil {
		return time.Location{}, fmt.Errorf("invalid Time-Zone header: %w", err)
	}

	return *location, nil
}
