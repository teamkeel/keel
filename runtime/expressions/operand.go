package expressions

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
)

// IsModelDbColumn returns true if the expression operand refers to a field value residing in the database.
// For example, a where condition might filter on reading data,
// such as: @where(post.author.isActive)
func IsModelDbColumn(model *proto.Model, fragments []string) bool {
	return fragments[0] == strcase.ToLowerCamel(model.Name)
}

// IsContextDbColumn returns true if the expression refers to a value on the context
// which will require database access (such as with identity backlinks),
// such as: @permission(expression: ctx.identity.user.isActive)
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
