package runtimectx

import "github.com/teamkeel/keel/proto"

const ContextTarget string = "ctx"

const (
	ContextIdentityField = "identity"
	ContextNowField      = "now"
)

var ContextFieldTypes = map[string]proto.Type{
	ContextIdentityField: proto.Type_TYPE_IDENTITY,
	ContextNowField:      proto.Type_TYPE_DATETIME,
}
