package actions

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/schema/parser"
)

// Include a filter (where condition) on the query based on an implicit input filter.
func (query *QueryBuilder) whereByImplicitFilter(scope *Scope, targetField []string, fieldName string, operator ActionOperator, value any) error {
	// Implicit inputs don't include the base model as the first fragment (unlike expressions), so we include it
	fragments := append([]string{casing.ToLowerCamel(scope.Action.ModelName)}, targetField...)

	// The lhs QueryOperand is determined from the fragments in the implicit input field
	left, err := operandFromFragments(scope.Schema, fragments)
	if err != nil {
		return err
	}

	// The rhs QueryOperand is always a value in an implicit input
	right := Value(value)

	// Add join for the implicit input
	err = query.addJoinFromFragments(scope, fragments)
	if err != nil {
		return err
	}

	// Add where condition to the query for the implicit input
	err = query.Where(left, operator, right)
	if err != nil {
		return err
	}

	return nil
}

// Include a filter (where condition) on the query based on an expression.
func (query *QueryBuilder) whereByExpression(scope *Scope, expression *parser.Expression, args map[string]any) error {
	// Only use parenthesis if there are multiple conditions
	useParenthesis := len(expression.Or) > 1
	for _, or := range expression.Or {
		if len(or.And) > 1 {
			useParenthesis = true
			break
		}
	}

	if useParenthesis {
		query.OpenParenthesis()
	}

	for _, or := range expression.Or {
		for _, and := range or.And {
			if and.Expression != nil {
				err := query.whereByExpression(scope, and.Expression, args)
				if err != nil {
					return err
				}
			}

			if and.Condition != nil {
				err := query.whereByCondition(scope, and.Condition, args)
				if err != nil {
					return err
				}
			}
			query.And()
		}
		query.Or()
	}

	if useParenthesis {
		query.CloseParenthesis()
	}

	return nil
}

// Include a filter (where condition) on the query based on a single condition.
func (query *QueryBuilder) whereByCondition(scope *Scope, condition *parser.Condition, args map[string]any) error {
	if condition.Type() != parser.ValueCondition && condition.Type() != parser.LogicalCondition {
		return fmt.Errorf("can only handle condition type of LogicalCondition or ValueCondition, have: %s", condition.Type())
	}

	lhsResolver := expressions.NewOperandResolver(scope.Context, scope.Schema, scope.Model, scope.Action, condition.LHS)
	rhsResolver := expressions.NewOperandResolver(scope.Context, scope.Schema, scope.Model, scope.Action, condition.RHS)

	lhsOperandType, err := lhsResolver.GetOperandType()
	if err != nil {
		return fmt.Errorf("cannot resolve operand type of LHS operand")
	}

	var operator ActionOperator
	var left, right *QueryOperand

	// Generate lhs QueryOperand
	left, err = generateQueryOperand(lhsResolver, args)
	if err != nil {
		return err
	}

	if lhsResolver.IsDatabaseColumn() {
		lhsFragments := lo.Map(lhsResolver.Operand.Ident.Fragments, func(fragment *parser.IdentFragment, _ int) string { return fragment.Fragment })

		// Generates joins based on the fragments that make up the operand
		err = query.addJoinFromFragments(scope, lhsFragments)
		if err != nil {
			return err
		}
	}

	if condition.Type() == parser.ValueCondition {
		if lhsOperandType != proto.Type_TYPE_BOOL {
			return fmt.Errorf("single operands in a value condition must be of type boolean")
		}

		// A value condition only has one operand in the expression,
		// for example, permission(expression: ctx.isAuthenticated),
		// so we must set the operator and RHS value (== true) ourselves.
		operator = Equals
		right = Value(true)
	} else {
		// The operator used in the expression
		operator, err = expressionOperatorToActionOperator(condition.Operator.ToString())
		if err != nil {
			return err
		}

		// Generate the rhs QueryOperand
		right, err = generateQueryOperand(rhsResolver, args)
		if err != nil {
			return err
		}

		if rhsResolver.IsDatabaseColumn() {
			rhsFragments := lo.Map(rhsResolver.Operand.Ident.Fragments, func(fragment *parser.IdentFragment, _ int) string { return fragment.Fragment })

			// Generates joins based on the fragments that make up the operand
			err = query.addJoinFromFragments(scope, rhsFragments)
			if err != nil {
				return err
			}
		}
	}

	// Adds where condition to the query for the expression
	err = query.Where(left, operator, right)
	if err != nil {
		return err
	}

	return nil
}

// Constructs and adds an INNER JOIN from a splice of fragments (representing an operand in an expression or implicit input).
// The fragment slice must include the base model as the first item, for example: "post." in post.author.publisher.isActive
func (query *QueryBuilder) addJoinFromFragments(scope *Scope, fragments []string) error {
	model := casing.ToCamel(fragments[0])
	fragmentCount := len(fragments)
	//previousIsRepeated := false

	for i := 1; i < fragmentCount-1; i++ {
		currentFragment := fragments[i]

		if !proto.ModelHasField(scope.Schema, model, currentFragment) {
			return fmt.Errorf("this model: %s, does not have a field of name: %s", model, currentFragment)
		}

		// We know that the current fragment is a related model because it's not the last fragment
		relatedModelField := proto.FindField(scope.Schema.Models, model, currentFragment)
		relatedModel := relatedModelField.Type.ModelName.Value
		foreignKeyField := proto.GetForignKeyFieldName(scope.Schema.Models, relatedModelField)
		primaryKey := "id"

		var leftOperand *QueryOperand
		var rightOperand *QueryOperand

		if proto.IsBelongsTo(relatedModelField) {
			// In a "belongs to" the foriegn key is on _this_ model
			leftOperand = ExpressionField(fragments[:i+1], primaryKey)
			rightOperand = ExpressionField(fragments[:i], foreignKeyField)
		} else {
			// In all others the foriegn key is on the _other_ model
			leftOperand = ExpressionField(fragments[:i+1], foreignKeyField)
			rightOperand = ExpressionField(fragments[:i], primaryKey)

		}

		query.Join(relatedModel, leftOperand, rightOperand)

		model = relatedModelField.Type.ModelName.Value
	}

	return nil
}

// Constructs a QueryOperand from a splice of fragments, representing an expression operand or implicit input.
// The fragment slice must include the base model as the first fragment, for example: post.author.publisher.isActive
func operandFromFragments(schema *proto.Schema, fragments []string) (*QueryOperand, error) {
	var field string
	model := casing.ToCamel(fragments[0])
	fragmentCount := len(fragments)

	for i := 1; i < fragmentCount; i++ {
		currentFragment := fragments[i]

		if !proto.ModelHasField(schema, model, currentFragment) {
			return nil, fmt.Errorf("this model: %s, does not have a field of name: %s", model, currentFragment)
		}

		if i < fragmentCount-1 {
			// We know that the current fragment is a model because it's not the last fragment
			relatedModelField := proto.FindField(schema.Models, model, currentFragment)
			model = relatedModelField.Type.ModelName.Value
		} else {
			// The last fragment is referencing the field
			field = currentFragment
		}
	}

	return ExpressionField(fragments[:len(fragments)-1], field), nil
}

// Generates a database QueryOperand, either representing a field, a value or null.
func generateQueryOperand(resolver *expressions.OperandResolver, args map[string]any) (*QueryOperand, error) {
	var queryOperand *QueryOperand

	if !resolver.IsDatabaseColumn() {
		value, err := resolver.ResolveValue(args)
		if err != nil {
			return nil, err
		}

		if value == nil {
			queryOperand = Null()
		} else {
			queryOperand = Value(value)
		}
	} else {
		// Step through the fragments in order to determine the table and field referenced by the expression operand
		fragments := lo.Map(resolver.Operand.Ident.Fragments, func(fragment *parser.IdentFragment, _ int) string { return fragment.Fragment })

		operandType, err := resolver.GetOperandType()
		if err != nil {
			return nil, err
		}

		// If the target is type MODEL, then refer to the
		// foreign key id by appending "Id" to the field name
		if operandType == proto.Type_TYPE_MODEL {
			fragments[len(fragments)-1] = fmt.Sprintf("%sId", fragments[len(fragments)-1])
		}

		// Generate QueryOperand from the fragments that make up the expression operand
		queryOperand, err = operandFromFragments(resolver.Schema, fragments)
		if err != nil {
			return nil, err
		}
	}

	return queryOperand, nil
}
