package expressions

import (
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

// IsModelDbColumn returns true if the expression operand refers to a field value residing in the database.
// For example, a where condition might filter on reading data,
// such as: @where(post.author.isActive).
func IsModelDbColumn(model *proto.Model, fragments []string) bool {
	return fragments[0] == strcase.ToLowerCamel(model.GetName())
}

// IsContextDbColumn returns true if the expression refers to a value on the context
// which will require database access (such as with identity backlinks),
// such as: @permission(expression: ctx.identity.user.isActive).
func IsContextDbColumn(fragments []string) bool {
	return IsContextIdentity(fragments) && !IsContextIdentityId(fragments)
}

func IsContextIdentity(fragments []string) bool {
	if !IsContext(fragments) {
		return false
	}
	if len(fragments) > 1 && fragments[1] == "identity" {
		return true
	}

	return false
}

func IsContextIdentityId(fragments []string) bool {
	if !IsContextIdentity(fragments) {
		return false
	}
	if len(fragments) == 2 {
		return true
	}
	if len(fragments) == 3 && fragments[2] == "id" {
		return true
	}

	return false
}

func IsContextIsAuthenticatedField(fragments []string) bool {
	if IsContext(fragments) && len(fragments) == 2 {
		return fragments[1] == "isAuthenticated"
	}

	return false
}

func IsContextField(fragments []string) bool {
	return IsContext(fragments) && !IsContextDbColumn(fragments)
}

func IsContextNowField(fragments []string) bool {
	if IsContext(fragments) && len(fragments) == 2 {
		return fragments[1] == "now"
	}
	return false
}

func IsContextHeadersField(fragments []string) bool {
	if IsContext(fragments) && len(fragments) == 3 {
		return fragments[1] == "headers"
	}
	return false
}

func IsContextEnvField(fragments []string) bool {
	if IsContext(fragments) && len(fragments) == 3 {
		return fragments[1] == "env"
	}
	return false
}

func IsContextSecretField(fragments []string) bool {
	if IsContext(fragments) && len(fragments) == 3 {
		return fragments[1] == "secrets"
	}
	return false
}

func IsContext(fragments []string) bool {
	return fragments[0] == "ctx"
}

// IsImplicitInput returns true if the expression operand refers to an implicit input on an action.
// For example, an input value provided in a create action might require validation,
// such as: create createThing() with (name) @validation(name != "").
func IsImplicitInput(schema *proto.Schema, action *proto.Action, fragments []string) bool {
	if len(fragments) <= 1 {
		return false
	}

	foundImplicitWhereInput := false
	foundImplicitValueInput := false

	whereInputs := proto.FindWhereInputMessage(schema, action.GetName())
	if whereInputs != nil {
		_, foundImplicitWhereInput = lo.Find(whereInputs.GetFields(), func(in *proto.MessageField) bool {
			return in.GetName() == fragments[0] && in.IsModelField()
		})
	}

	valuesInputs := proto.FindValuesInputMessage(schema, action.GetName())
	if valuesInputs != nil {
		_, foundImplicitValueInput = lo.Find(valuesInputs.GetFields(), func(in *proto.MessageField) bool {
			return in.GetName() == fragments[0] && in.IsModelField()
		})
	}

	return foundImplicitWhereInput || foundImplicitValueInput
}

// IsInput returns true if the expression operand refers to a named input or a model field input on an action.
// For example, for a where condition might use an named input,
// such as: list listThings(isActive: Boolean) @where(thing.isActive == isActive)
// Or a model field input,
// such as: list listThings(thing.isActive).
func IsInput(schema *proto.Schema, action *proto.Action, fragments []string) bool {
	foundExplicitWhereInput := false
	foundExplicitValueInput := false

	whereInputs := proto.FindWhereInputMessage(schema, action.GetName())
	if whereInputs != nil {
		_, foundExplicitWhereInput = lo.Find(whereInputs.GetFields(), func(in *proto.MessageField) bool {
			return in.GetName() == fragments[0]
		})
	}

	valuesInputs := proto.FindValuesInputMessage(schema, action.GetName())
	if valuesInputs != nil {
		_, foundExplicitValueInput = lo.Find(valuesInputs.GetFields(), func(in *proto.MessageField) bool {
			return in.GetName() == fragments[0]
		})
	}

	return foundExplicitWhereInput || foundExplicitValueInput
}
