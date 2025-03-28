package flows

import (
	"github.com/teamkeel/keel/proto"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/flows")

type Scope struct {
	Flow   *proto.Flow
	Schema *proto.Schema
}

func NewScope(flow *proto.Flow, schema *proto.Schema) *Scope {
	return &Scope{
		Flow:   flow,
		Schema: schema,
	}
}
