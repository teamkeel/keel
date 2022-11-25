package runtimectx

import "github.com/teamkeel/keel/proto"

const ContextTarget string = "ctx"

const (
	ContextIdentityField        = "identity"
	ContextIsAuthenticatedField = "isAuthenticated"
	ContextNowField             = "now"
)

var ContextFieldTypes = map[string]proto.Type{
	ContextIdentityField:        proto.Type_TYPE_MODEL,
	ContextIsAuthenticatedField: proto.Type_TYPE_BOOL,
	ContextNowField:             proto.Type_TYPE_DATETIME,
}
