package validation

import (
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
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
	// actions.ActionTypesRule,
	// actions.ValidActionInputTypesRule,
	// actions.ValidActionInputLabelRule,
	// actions.ValidArbitraryFunctionReturns,
	// actions.ActionModelInputsRule,
	// actions.CreateOperationNoReadInputsRule,
	// actions.CreateOperationRequiredFieldsRule,

	// field.ValidFieldTypesRule,
	// field.UniqueFieldNamesRule,
	// field.FieldNamesMaxLengthRule,

	// model.ModelNamesMaxLengthRule,

	// attribute.AttributeLocationsRule,
	// attribute.SetWhereAttributeRule,
	// attribute.ValidateFieldAttributeRule,
	// attribute.UniqueAttributeArgsRule,

	// role.UniqueRoleNamesRule,

	// api.UniqueAPINamesRule,
	// api.NamesCorrespondToModels,
}

var visitorFuncs = []VisitorFunc{
	// DuplicateModelNames,
	// DuplicateEnumNames,
	// DuplicateMessageNames,
	// DuplicateActionNames,
	// DuplicateJobNames,
	// CasingRule,
	// NameClashesRule,
	// RecursiveFieldsRule,
	// RequiredFieldOfSameModelType,
	// UnusedInputRule,
	// NotMutableInputs,
	// CreateNestedInputIsMany,
	// ConflictingInputsRule,
	// UniqueLookup,
	// InvalidWithUsage,
	// UniqueAttributeRule,
	// OrderByAttributeRule,
	// SortableAttributeRule,
	// SetAttributeExpressionRules,
	// Jobs,
	// MessagesRule,
	// ScheduleAttributeRule,
	// DuplicateInputsRule,
	// PermissionsAttributeArguments,
	// FunctionDisallowedBehavioursRule,
	// OnAttributeRule,
	// EmbedAttributeRule,
	// RelationshipsRules,
	// ApiModelActions,
	// StudioFeatures,
}

// RunAllValidators will run all the validators available. If withWarnings is true, it will return the errors even if
// they contain just warnings
func (v *Validator) RunAllValidators(withWarnings bool) (errs *errorhandling.ValidationErrors) {
	errs = &errorhandling.ValidationErrors{}

	for _, vf := range validatorFuncs {
		errs.Concat(vf(v.asts))
	}

	visitors := []Visitor{}
	for _, fn := range visitorFuncs {
		visitors = append(visitors, fn(v.asts, errs))
	}

	runVisitors(v.asts, visitors)

	// if we've got any warnings and they should be included, just return, no need to check for actual errors
	if withWarnings && len(errs.Warnings) > 0 {
		return errs
	}

	if len(errs.Errors) == 0 {
		return nil
	}

	return errs
}
