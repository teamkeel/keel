package runtimectx

import "github.com/teamkeel/keel/proto"

type contextKey string

const ContextTarget string = "ctx"

const (
	ContextIdentityField        = "identity"
	ContextIsAuthenticatedField = "isAuthenticated"
	ContextNowField             = "now"
	ContextEnvField             = "env"
	ContextSecretField          = "secret"
)

var ContextFieldTypes = map[string]proto.Type{
	ContextIdentityField:        proto.Type_TYPE_MODEL,
	ContextIsAuthenticatedField: proto.Type_TYPE_BOOL,
	ContextNowField:             proto.Type_TYPE_DATETIME,
	ContextEnvField:             proto.Type_TYPE_OBJECT,
	ContextSecretField:          proto.Type_TYPE_SECRET,
}
