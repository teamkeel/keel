package flows

import (
	"context"

	"github.com/teamkeel/keel/proto"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/flows")

type Scope struct {
	Context context.Context
	Flow    *proto.Flow
	Schema  *proto.Schema
}

// WithContext sets the given context on the scope
func (s *Scope) WithContext(ctx context.Context) *Scope {
	return &Scope{
		Context: ctx,
		Flow:    s.Flow,
		Schema:  s.Schema,
	}
}

func NewScope(ctx context.Context, flow *proto.Flow, schema *proto.Schema) *Scope {
	return &Scope{
		Context: ctx,
		Flow:    flow,
		Schema:  schema,
	}
}
