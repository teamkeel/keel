package actions

import (
	"fmt"
	"strings"

	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel/trace"
)

// RegisterAuditScopeInDB extracts the Identity associated from the given
// Scope (if available), and the Span Id from the given Span,
// and posts them into the database cited by the scope.
//
// The aim is to make these data available to the process_audit() function that fires
// inside Postgres when rows are mutated.
//
// When these data are missing from the scope, it uses the constant string "missing".
func RegisterAuditScopeInDB(scope *Scope, span trace.Span) (err error) {

	// Capture the required data from the scope and the trace span.

	ctx := scope.Context

	var identityId string = "missing"
	var traceSpanId string = "missing"

	if identity, err := runtimectx.GetIdentity(ctx); err == nil {
		identityId = identity.Id
	}

	if span.SpanContext().HasSpanID() {
		traceSpanId = span.SpanContext().SpanID().String()
		// XXXX remove this
		fmt.Printf("XXXX span has Id, which is <%s>\n", traceSpanId)
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
