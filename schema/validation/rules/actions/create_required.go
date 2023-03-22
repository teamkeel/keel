package actions

import (
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// CreateOperationRequiredFieldsRule makes sure that all create operation are specified in such a way
// that all the fields that must be populated during a create, are covered by either
// inputs or set expressions.
// This includes (recursively) the fields in nested models where appropriate.
func CreateOperationRequiredFieldsRule(
	asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		rootModelName := strcase.ToLowerCamel(model.Name.Value)
		for _, op := range query.ModelCreateOperations(model) {
			dotDelimPath := ""
			for _, field := range query.ModelFields(model) {
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
// by definition. For example, if it is optional etc.
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
			errorhandling.ErrorCreateOperationMissingInput,
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

	// Is the operation REFERENCING an existing related model instance, or
	// is it CREATING the related model instance?

	// e.g. author.pet.id

	pathForReferencing := extendDotDelimPath(dotDelimPath, field.Name.Value)
	pathForReferencing = extendDotDelimPath(pathForReferencing, parser.ImplicitFieldNameId)

	// If it is referencing, then the fields inside the reference model are not expected
	// for this Create op.
	if satisfied(rootModelName, pathForReferencing, model.Name.Value, op) {
		return
	}

	// So it is creating the related model... We must recurse to make sure the
	// those of the related model's fields that are required for creation are
	// provided as inputs.

	nestedModel := query.Model(asts, field.Type)
	nestedPath := extendDotDelimPath(dotDelimPath, field.Name.Value)
	for _, nestedModelField := range query.ModelFields(nestedModel) {

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
