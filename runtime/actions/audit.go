package actions

import (
	"fmt"
	"strings"

	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel/trace"
)

// RegisterAuditScopeInDB extracts the Identity and Trace Id associated from the given
// Scope (if available), and posts them into the database cited by the scope.
//
// The aim is to make these data available to the process_audit() function that fires
// inside Postgres when rows are mutated.
//
// When these data are missing from the scope, it uses the constant strings:
// MissingIdentity and MissingTraceId respectively.
func RegisterAuditScopeInDB(scope *Scope, span trace.Span) (err error) {

	// Capture the required data from the scope and the trace span.

	ctx := scope.Context

	// XXXX remove this debug
	hasSpanID := span.SpanContext().HasSpanID()
	hasTraceID := span.SpanContext().HasTraceID()

	_, _ = hasSpanID, hasTraceID

	if hasSpanID {
		fmt.Printf("XXXX found case where has span id is set\n")
	}
	if hasTraceID {
		fmt.Printf("XXXX found case where has span id is set\n")
	}

	var identityId string = MissingIdentity
	var traceSpanId string = MissingTraceId

	if identity, err := runtimectx.GetIdentity(ctx); err == nil {
		identityId = identity.Id
	}

	if span.SpanContext().HasSpanID() {
		traceSpanId = span.SpanContext().SpanID().String()
	}

	// Write the captured data into postgres config.
	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return err
	}

	// The 'false' in the SQL below is the "local" argument for setting config values. Local=true means
	// transactions scope, whereas local=false means connection scope).
	s1 := fmt.Sprintf("select set_config('audit.identity_id', '%s', false);", identityId)
	s2 := fmt.Sprintf("select set_config('audit.trace_id', '%s', false);", traceSpanId)
	sql := strings.Join([]string{s1, s2}, "\n")
	_, err = db.ExecuteStatement(ctx, sql)
	if err != nil {
		return err
	}
	return nil
}

const MissingIdentity = "missing_identity"
const MissingTraceId = "missing_trace_id"
