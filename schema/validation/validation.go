package validation

import (
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/schema/validation/rules/actions"
	"github.com/teamkeel/keel/schema/validation/rules/api"
	"github.com/teamkeel/keel/schema/validation/rules/attribute"
	"github.com/teamkeel/keel/schema/validation/rules/field"
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

type validationFunc func(asts []*parser.AST) errorhandling.ValidationErrors

var validatorFuncs = []validationFunc{
	actions.ActionTypesRule,
	actions.UniqueActionNamesRule,
	actions.ValidActionInputTypesRule,
	actions.ValidActionInputLabelRule,
	actions.ValidArbitraryFunctionReturns,
	actions.ActionModelInputsRule,
	actions.CreateOperationNoReadInputsRule,
	actions.CreateOperationRequiredFieldsRule,
	actions.ReservedActionNameRule,

	field.ValidFieldTypesRule,
	field.UniqueFieldNamesRule,

	attribute.AttributeLocationsRule,
	attribute.SetWhereAttributeRule,
	attribute.ValidateFieldAttributeRule,
	attribute.UniqueAttributeArgsRule,

	role.UniqueRoleNamesRule,

	api.UniqueAPINamesRule,
	api.NamesCorrespondToModels,
}

var visitorFuncs = []VisitorFunc{
	DuplicateDefinitionRule,
	CasingRule,
	NameClashesRule,
	RecursiveFieldsRule,
	RequiredFieldOfSameModelType,
	UnusedInputRule,
	NotMutableInputs,
	CreateNestedInputIsMany,
	ConflictingInputsRule,
	UniqueLookup,
	InvalidWithUsage,
	RepeatedScalarFieldRule,
	UniqueAttributeRule,
	OrderByAttributeRule,
	SortableAttributeRule,
	SetAttributeExpressionRules,
	Jobs,
	MessagesRule,
	ScheduleAttributeRule,
	DuplicateInputsRule,
	PermissionsAttributeArguments,
	FunctionDisallowedBehavioursRule,
	OnAttributeRule,
	RelationshipsRules,
}

func (v *Validator) RunAllValidators() (errs *errorhandling.ValidationErrors) {
	errs = &errorhandling.ValidationErrors{}

	for _, vf := range validatorFuncs {
		errs.Concat(vf(v.asts))
	}

	visitors := []Visitor{}
	for _, fn := range visitorFuncs {
		visitors = append(visitors, fn(v.asts, errs))
	}

	runVisitors(v.asts, visitors)

	if len(errs.Errors) == 0 {
		return nil
	}

	return errs
}
