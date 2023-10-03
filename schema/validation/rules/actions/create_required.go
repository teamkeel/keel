package actions

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// CreateOperationRequiredFieldsRule makes sure that all create operation are specified in such a way
// that all the fields that must be populated during a create, are covered by either
// inputs or set expressions.
// Only applies to built in actions
// This includes (recursively) the fields in nested models where appropriate.
func CreateOperationRequiredFieldsRule(
	asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		rootModelName := casing.ToLowerCamel(model.Name.Value)

		for _, op := range query.ModelCreateActions(model, func(a *parser.ActionNode) bool { return !a.IsFunction() }) {
			dotDelimPath := ""
			for _, field := range query.ModelFields(model) {
				if field.Type.Value == model.Name.Value && !field.Optional {
					// This avoids an infinite loop in the case that the field type is the same as the model type,
					// and is not optional, which is not a valid schema and is handled elsewhere in schema validations.
					continue
				}
				checkField(asts, field, model, rootModelName, dotDelimPath, op, &errs)
			}
		}
	}
	return errs
}

// checkField makes sure that if the given field is mandatory for the given operation,
// that it is provided by an operation input or @set expression.
// It checks recursively for nested related models where appropriate.
func checkField(
	asts []*parser.AST,
	field *parser.FieldNode,
	model *parser.ModelNode, // Model the field belongs to.
	rootModelName string, // the "post" in @set(post.foo.bar.baz.id = something)
	dotDelimPath string, // the "foo.bar.baz.id" in @set(post.foo.bar.baz.id = something)
	op *parser.ActionNode,
	errs *errorhandling.ValidationErrors,
) {
	if isNotNeeded(field) {
		return
	}
	switch {
	case query.IsHasOneModelField(asts, field):
		checkHasOneRelationField(asts, field, model, rootModelName, dotDelimPath, op, errs)
	default:
		checkPlainField(asts, field, model, rootModelName, dotDelimPath, op, errs)
	}
}

// isNotNeeded works out if the given field is not needed for a create operation
// by definition. This IS the case for:
// - optional fields
// - repeated fields
// - fields which have a default
// - built-in fields like CreatedAt, Id etc.
func isNotNeeded(f *parser.FieldNode) bool {
	switch {
	case f.Optional, f.Repeated, query.FieldHasAttribute(f, parser.AttributeDefault), f.BuiltIn:
		return true
	default:
		return false
	}
}

// checkPlainField works out if the given (non-relationship) field is set by either one of the
// given operation's inputs, or one of its @set expressions.
//
// see checkField() for comments on the arguments.
func checkPlainField(
	asts []*parser.AST,
	field *parser.FieldNode,
	model *parser.ModelNode,
	rootModelName string,
	dotDelimPath string,
	op *parser.ActionNode,
	errs *errorhandling.ValidationErrors,
) {

	requiredPath := extendDotDelimPath(dotDelimPath, field.Name.Value)

	if !satisfied(rootModelName, requiredPath, model.Name.Value, op) {
		errs.Append(
			errorhandling.ErrorCreateActionMissingInput,
			map[string]string{
				"FieldName": requiredPath,
			},
			op,
		)
	}
}

// checkHasOneRelationField looks works out if the given (has-one-relationship) field is satisfied
// by either one of the given operation's inputs, or by one of its @set expressions.
//
// The field can be satisfied EITHER with an input that references an EXISTING instance of the related model
// using the form "author.pet.id". Or with ALL of the mandatory fields of the related model
// being satisfied (recursively).
//
// see checkField() for comments on the arguments.
func checkHasOneRelationField(
	asts []*parser.AST,
	field *parser.FieldNode,
	model *parser.ModelNode,
	rootModelName string,
	dotDelimPath string,
	op *parser.ActionNode,
	errs *errorhandling.ValidationErrors,
) {

	nestedModel := query.Model(asts, field.Type.Value)
	pathToReferencedModel := extendDotDelimPath(dotDelimPath, field.Name.Value)
	pathToReferencedModelDotID := extendDotDelimPath(pathToReferencedModel, parser.ImplicitFieldNameId)

	// The field itself can be set in a @set expression. An example of this is identity e.g.
	//   @set(myModel.identityField = ctx.identity)
	fieldIsSet := satisfiedBySetExpr(rootModelName, pathToReferencedModel, model.Name.Value, op)

	// The id field of the relation can also be set in either a @set or an input. For examoke:
	//   @set(myModel.myField.id = someValue)
	// or
	//   create myAction() with (myField.id)
	fieldIdIsSet := satisfied(rootModelName, pathToReferencedModelDotID, model.Name.Value, op)

	// If the field is being set to an existing record then we make sure no other fields on the model are being set.
	if fieldIsSet || fieldIdIsSet {
		makeSureReferencedFieldsAreNotGiven(asts, nestedModel, rootModelName, pathToReferencedModel, op, errs)
		return
	}

	// Special case to improve error message for Identity fields
	if nestedModel.Name.Value == parser.ImplicitIdentityModelName {
		message := fmt.Sprintf("the %s field of %s is not set as part of this create operation", field.Name.Value, model.Name.Value)
		errs.AppendError(
			errorhandling.NewValidationErrorWithDetails(
				errorhandling.ActionInputError,
				errorhandling.ErrorDetails{
					Message: message,
					Hint:    fmt.Sprintf("set the field using: @set(%s.%s = ctx.identity)", rootModelName, pathToReferencedModel),
				},
				op.Name),
		)
		return
	}

	// We have established that the operation does intend to create this nested model instance.
	// Therefore we must recurse to make sure the creation required fields for the nested model are
	// supplied.
	nestedPath := extendDotDelimPath(dotDelimPath, field.Name.Value)
	for _, nestedModelField := range query.ModelFields(nestedModel) {
		// Skip if the field is the other side of a 1:1 relationship.
		// TODO: Support multiple 1:1 relationships between the same two tables.
		if nestedModelField.Name.Value == rootModelName && !nestedModelField.Repeated {
			continue
		}
		// This is where the recursion happens.
		checkField(asts, nestedModelField, nestedModel, rootModelName, nestedPath, op, errs)
	}
}

// Satisfied returns true if the given requiredField (including dotted path where appropriate),
// is set either by a with() clause on the operation, or by one of its @set expressions.
//
// see checkField() for comments on the arguments.
func satisfied(rootModelName string, requiredField string, modelName string, op *parser.ActionNode) bool {
	if requiredFieldInWithClause(requiredField, op) {
		return true
	}
	if satisfiedBySetExpr(rootModelName, requiredField, modelName, op) {
		return true
	}
	return false
}

// makeSureReferencedFieldsAreNotGiven creates validation errors if the given operation has any
// inputs, or @set expressions that target fields of the given referenced model using the given dotted
// path ancestry. (apart from the one that ends in ".id")
func makeSureReferencedFieldsAreNotGiven(
	asts []*parser.AST,
	referencedModel *parser.ModelNode,
	rootModelName string,
	pathToReferencedModel string,
	op *parser.ActionNode,
	errs *errorhandling.ValidationErrors) {

	referencedModelName := casing.ToLowerCamel(referencedModel.Name.Value)
	pathToIgnore := extendDotDelimPath(pathToReferencedModel, parser.ImplicitFieldNameId)
	for _, referencedField := range query.ModelFields(referencedModel) {
		pathToField := extendDotDelimPath(pathToReferencedModel, referencedField.Name.Value)
		if pathToField == pathToIgnore {
			continue
		}
		// Note we are making sure the path is NOT satisfied.
		if satisfied(rootModelName, pathToField, referencedModelName, op) {
			errs.Append(
				errorhandling.ErrorCreateActionAmbiguousRelationship,
				map[string]string{
					"IdPath":          pathToIgnore,
					"ConflictingPath": pathToField,
					"ModelName":       referencedModelName,
				},
				op,
			)
		}
	}
}

// setExpressions returns all the non-nil expressions from all
// the @set attributes on the given action.
func setExpressions(action *parser.ActionNode) []*parser.Expression {
	setters := lo.Filter(action.Attributes, func(a *parser.AttributeNode, _ int) bool {
		return a.Name.Value == parser.AttributeSet
	})
	expressions := []*parser.Expression{}
	for _, setAttr := range setters {
		if len(setAttr.Arguments) == 0 {
			continue
		}
		if setAttr.Arguments[0].Expression != nil {
			expressions = append(expressions, setAttr.Arguments[0].Expression)
		}
	}
	return expressions
}

// requiredFieldInWithClause returns true if the given requiredField is
// present the the given action's "With" inputs.
func requiredFieldInWithClause(requiredField string, action *parser.ActionNode) bool {
	for _, input := range action.With {
		if input.Label == nil && input.Type.ToString() == requiredField {
			return true
		}
	}
	return false
}

// satisfiedBySetExpr works out if any of the operation's @set expressions are matching assignments
// with a LHS of this pattern: "author.pet.name".
//
// In order to match:
// - the first fragment of the LHS must be the given rootModelName (e.g. "author")
// - the remaining fragments - when joined with a dot, equal the remainder (e.g. "pet.name").
// It copes with an arbitrary number of fragments.
//
// see checkField() for comments on the arguments.
func satisfiedBySetExpr(rootModelName string, dotDelimPath string, modelName string, action *parser.ActionNode) bool {
	setExpressions := setExpressions(action)

	for _, expr := range setExpressions {

		assignment, err := expr.ToAssignmentCondition()
		if err != nil {
			continue
		}
		lhs := assignment.LHS

		fragStrings := lo.Map(lhs.Ident.Fragments, func(frag *parser.IdentFragment, _ int) string {
			return frag.Fragment
		})
		if len(fragStrings) < 2 {
			continue
		}

		if fragStrings[0] != rootModelName {
			continue
		}

		remainingFragments := fragStrings[1:]
		remainingPath := strings.Join(remainingFragments, ".")
		if remainingPath == dotDelimPath {
			return true
		}
	}
	return false
}

// extendDotDelimPath extends the given dot-delimited input path by adding
// a new dotted segment on the end - coping with the singularity
// of the input path being empty and not wanting a leading dot.
func extendDotDelimPath(inPath string, newSegment string) (outPath string) {
	if inPath == "" {
		outPath = newSegment
	} else {
		outPath = inPath + "." + newSegment
	}

	return outPath
}
