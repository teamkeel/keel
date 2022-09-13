package actions

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/expressions"
)

// interpretExpressionGivenArgs examines the given expression, in order to work out how to construct a gorm WHERE clause.
//
// The ONLY form we support at the moment in this infant version is this: "person.name == name"
//
// It returns a column and a value that is good to be used like this:
//
//	tx.Where(fmt.Sprintf("%s = ?", column), value)
func interpretExpressionGivenArgs(
	expr *expressions.Expression,
	operation *proto.Operation,
	schema *proto.Schema,
	args map[string]any,
) (column string, value any, err error) {

	// Make sure the expression is in the form we can handle.

	conditions := expr.Conditions()
	if len(conditions) != 1 {
		return "", nil, fmt.Errorf("cannot yet handle multiple conditions, have: %d", len(conditions))
	}
	condition := conditions[0]
	cType := condition.Type()
	if cType != expressions.LogicalCondition {
		return "", nil, fmt.Errorf("cannot yet handle condition types other than LogicalCondition, have: %s", cType)
	}

	if condition.Operator.ToString() != expressions.OperatorEquals {
		return "", nil, fmt.Errorf(
			"cannot yet handle operators other than OperatorEquals, have: %s",
			condition.Operator.ToString())
	}

	if condition.LHS.Type() != expressions.TypeIdent {
		return "", nil, fmt.Errorf("cannot handle LHS of type other than TypeIdent, have: %s", condition.LHS.Type())
	}
	if condition.RHS.Type() != expressions.TypeIdent {
		return "", nil, fmt.Errorf("cannot handle RHS of type other than TypeIdent, have: %s", condition.LHS.Type())
	}

	lhs := condition.LHS
	if len(lhs.Ident.Fragments) != 2 {
		return "", nil, fmt.Errorf("cannot handle LHS identifier unless it has 2 fragments, have: %d", len(lhs.Ident.Fragments))
	}

	rhs := condition.RHS
	if len(rhs.Ident.Fragments) != 1 {
		return "", nil, fmt.Errorf("cannot handle RHS identifier unless it has 1 fragment, have: %d", len(rhs.Ident.Fragments))
	}

	// Make sure the first fragment in the LHS is the name of the model of which this operation is part.
	// e.g. "person" in the example above.
	modelTarget := strcase.ToCamel(lhs.Ident.Fragments[0].Fragment)
	if modelTarget != operation.ModelName {
		return "", nil, fmt.Errorf("can only handle the first LHS fragment referencing the Operation's model, have: %s", modelTarget)
	}

	// Make sure the second fragment in the LHS is the name of a field of the model of which this operation is part.
	// e.g. "name" in the example above.
	fieldName := lhs.Ident.Fragments[1].Fragment
	if !proto.ModelHasField(schema, modelTarget, fieldName) {
		return "", nil, fmt.Errorf("this model: %s, does not have a field of name: %s", modelTarget, fieldName)
	}

	// Make sure the single fragment in the RHS matches up with an expected Input for this operation.
	inputName := rhs.Ident.Fragments[0].Fragment
	if !proto.OperationHasInput(operation, inputName) {
		return "", nil, fmt.Errorf("operation does not define an input called: %s", inputName)
	}

	// Make sure the "where" part of the args on the specified input has been provided in the given Args
	arg, ok := args[inputName]
	if !ok {
		return "", nil, fmt.Errorf("request does not have provide argument of name: %s", inputName)
	}

	// Now we have all the data we need to return
	return strcase.ToSnake(fieldName), arg, nil
}

// TODO - need to DRY up and rationalise the functions above and below!!!

// interpretExpressionField examines the given expression, in order to work out how to construct a gorm WHERE clause.
//
// The ONLY form we support at the moment in this infant version is this: "person.name == <an-input-name>"
//
// It returns the field name being assigned to.
//
//	tx.Where(fmt.Sprintf("%s = ?", column), value)
func interpretExpressionField(
	expr *expressions.Expression,
	operation *proto.Operation,
	schema *proto.Schema,
) (column string, err error) {

	// Make sure the expression is in the form we can handle.

	conditions := expr.Conditions()
	if len(conditions) != 1 {
		return "", fmt.Errorf("cannot yet handle multiple conditions, have: %d", len(conditions))
	}
	condition := conditions[0]
	cType := condition.Type()
	if cType != expressions.LogicalCondition {
		return "", fmt.Errorf("cannot yet handle condition types other than LogicalCondition, have: %s", cType)
	}

	if condition.Operator.ToString() != expressions.OperatorEquals {
		return "", fmt.Errorf(
			"cannot yet handle operators other than OperatorEquals, have: %s",
			condition.Operator.ToString())
	}

	if condition.LHS.Type() != expressions.TypeIdent {
		return "", fmt.Errorf("cannot handle LHS of type other than TypeIdent, have: %s", condition.LHS.Type())
	}
	if condition.RHS.Type() != expressions.TypeIdent {
		return "", fmt.Errorf("cannot handle RHS of type other than TypeIdent, have: %s", condition.LHS.Type())
	}

	lhs := condition.LHS
	if len(lhs.Ident.Fragments) != 2 {
		return "", fmt.Errorf("cannot handle LHS identifier unless it has 2 fragments, have: %d", len(lhs.Ident.Fragments))
	}

	rhs := condition.RHS
	if len(rhs.Ident.Fragments) != 1 {
		return "", fmt.Errorf("cannot handle RHS identifier unless it has 1 fragment, have: %d", len(rhs.Ident.Fragments))
	}

	// Make sure the first fragment in the LHS is the name of the model of which this operation is part.
	// e.g. "person" in the example above.
	modelTarget := strcase.ToCamel(lhs.Ident.Fragments[0].Fragment)
	if modelTarget != operation.ModelName {
		return "", fmt.Errorf("can only handle the first LHS fragment referencing the Operation's model, have: %s", modelTarget)
	}

	// Make sure the second fragment in the LHS is the name of a field of the model of which this operation is part.
	// e.g. "name" in the example above.
	fieldName := lhs.Ident.Fragments[1].Fragment
	if !proto.ModelHasField(schema, modelTarget, fieldName) {
		return "", fmt.Errorf("this model: %s, does not have a field of name: %s", modelTarget, fieldName)
	}

	// Now we have all the data we need to return
	return fieldName, nil
}
