package validation

import (
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/schema/validation/rules/api"
	"github.com/teamkeel/keel/schema/validation/rules/attribute"
	"github.com/teamkeel/keel/schema/validation/rules/enum"
	"github.com/teamkeel/keel/schema/validation/rules/field"
	"github.com/teamkeel/keel/schema/validation/rules/model"
	"github.com/teamkeel/keel/schema/validation/rules/role"
)

type Validator struct {
	asts []*parser.AST
}

func NewValidator(asts []*parser.AST) *Validator {
	return &Validator{
		asts: asts,
	}
}

// A Validator knows how to validate a parsed Keel schema.
//
// Conceptually we are validating a single schema.
// But the Validator supports it being "delivered" as a collection
// of *parser.Schema objects - to match up with a user's schema likely
// being written across N files.

type validationFunc func(asts []*parser.AST) []error

var validatorFuncs = []validationFunc{
	// Begin base model validations
	model.ReservedModelNamesRule,
	model.ModelNamingRule,
	model.UniqueModelNamesRule,
	// Begin sub actions of model
	model.ActionNamingRule,
	model.UniqueOperationNamesRule,
	model.ValidActionInputsRule,
	model.GetOperationUniqueLookupRule,
	// Begin fields
	field.ReservedNameRule,
	field.ValidFieldTypesRule,
	field.UniqueFieldNamesRule,
	field.FieldNamingRule,
	// Begin attribute validation
	attribute.AttributeLocationsRule,
	attribute.PermissionAttributeRule,
	attribute.SetAttributeRule,
	// Role
	role.UniqueRoleNamesRule,
	// API
	api.UniqueAPINamesRule,
	// Enum
	enum.UniqueEnumsRule,
}

func (v *Validator) RunAllValidators() error {
	var errors []*errorhandling.ValidationError

	for _, vf := range validatorFuncs {
		err := vf(v.asts)

		for _, e := range err {
			if verrs, ok := e.(*errorhandling.ValidationError); ok {
				errors = append(errors, verrs)
			}
		}
	}

	if len(errors) > 0 {
		errors := errorhandling.ValidationErrors{Errors: errors}
		return errors
	}

	return nil
}
