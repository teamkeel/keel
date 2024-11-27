package expressions

import (
	"context"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
)

type Act struct {
	context context.Context
	schema  *proto.Schema
	model   *proto.Model
	action  *proto.Action
}

func (a *Act) ResolveName(name string) (any, bool) {
	//resolver := NewOperandResolverCel()

	//expr

	switch name {
	case "ctx.isAuthenticated":

		return auth.IsAuthenticated(a.context), true
	case "ctx.identity", "ctx.identity.id":
		isAuthenticated := auth.IsAuthenticated(a.context)
		if !isAuthenticated {
			return false, true
		}

		identity, err := auth.GetIdentity(a.context)
		if err != nil {
			return false, false
		}

		return identity[parser.FieldNameId].(string), true
	case "ctx.now":
		return runtimectx.GetNow(), true
	}

	if secretName, found := strings.CutPrefix(name, "ctx.secrets."); found {
		secrets := runtimectx.GetSecrets(a.context)
		if value, ok := secrets[secretName]; ok {
			return value, true
		} else {
			return nil, true
		}
	}

	// if header, found := strings.CutPrefix(name, "ctx.headers."); found {
	// 	secrets := runtimectx.GetSecrets(a.context)
	// 	if value, ok := secrets[secretName]; ok {
	// 		return value, true
	// 	} else {
	// 		return nil, true
	// 	}
	// }

	return nil, false
}
