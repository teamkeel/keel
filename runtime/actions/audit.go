package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/runtime/auth"
	"go.opentelemetry.io/otel/trace"
)

// SetAuditScopeIntoDB writes some meta data about a request scope into
// the Postgres config. It writes values for Identity Id and Trace Id from the
// given scope and trace span respectively.
//
// If the values cannot be extracted from the ctx and
// the span, it sets the values to "not found".
func SetAuditScopeIntoDB(ctx context.Context, span trace.Span) (err error) {

	// Capture the required data from the scope and the trace span.

	identity, err := auth.GetIdentity(ctx)

	var identityId string = "not found"
	identityAvailable := (err == nil) && (identity != nil)
	if identityAvailable {
		identityId = identity.Id
		fmt.Printf("XXXX identity is available: %s\n", identityId)
	}

	var traceSpanId string = "not found"
	if span.SpanContext().HasSpanID() {
		traceSpanId = span.SpanContext().SpanID().String()
		fmt.Printf("XXXX traceSpanId is available: %s\n", traceSpanId)
	}

	// Write the captured data into postgres config.
	db, err := db.GetDatabase(ctx)
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

// ClearAuditScopeInDB is a sister function to SetAuditScopeIntoDB, but it
// sets the postgres config values to "unknown".
func ClearAuditScopeInDB(scope *Scope) {

	ctx := scope.Context

	db, err := db.GetDatabase(ctx)
	if err != nil {
		return
	}

	s1 := fmt.Sprintf("select set_config('audit.identity_id', '%s', false);", "unknown")
	s2 := fmt.Sprintf("select set_config('audit.trace_id', '%s', false);", "unknown")
	sql := strings.Join([]string{s1, s2}, "\n")
	_, err = db.ExecuteStatement(ctx, sql)
	if err != nil {
		return
	}
}
