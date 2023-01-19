package actions

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/schema/parser"
)

// Include a filter (where condition) on the query based on an implicit input filter.
func (query *QueryBuilder) whereByImplicitFilter(scope *Scope, input *proto.OperationInput, fieldName string, operator ActionOperator, value any) error {
	// Implicit inputs don't include the base model as the first fragment (unlike expressions), so we include it
	fragments := append([]string{strcase.ToLowerCamel(input.ModelName)}, input.Target...)

	// The lhs QueryOperand is determined from the fragments in the implicit input field
	left, _ := operandFromFragments(scope.schema, fragments)

	// The rhs QueryOperand is always a value in an implicit input
	right := Value(value)

	// Add join for the implicit input
	query.addJoinFromFragments(scope, fragments)

	// Add where condition to the query for the implicit input
	query.Where(left, operator, right)

	return nil
}

// Include a filter (where condition) on the query based on a filter expression.
func (query *QueryBuilder) whereByExpression(scope *Scope, expression *parser.Expression, args map[string]any) error {
	if len(expression.Conditions()) != 1 {
		return fmt.Errorf("cannot yet handle multiple conditions, have: %d", len(expression.Conditions()))
	}

	condition := expression.Conditions()[0]

	if condition.Type() != parser.ValueCondition && condition.Type() != parser.LogicalCondition {
		return fmt.Errorf("can only handle condition type of LogicalCondition or ValueCondition, have: %s", condition.Type())
	}

	lhsResolver := expressions.NewOperandResolver(scope.context, scope.schema, scope.operation, condition.LHS)
	rhsResolver := expressions.NewOperandResolver(scope.context, scope.schema, scope.operation, condition.RHS)

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
		query.addJoinFromFragments(scope, lhsFragments)
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
			query.addJoinFromFragments(scope, rhsFragments)
			if err != nil {
				return err
			}
		}
	}

	// Adds where condition to the query for the expression
	query.Where(left, operator, right)

	return nil
}

// Constructs and adds an INNER JOIN from a splice of fragments (representing an operand in an expression or implicit input).
// The fragment slice must include the base model as the first item, for example: "post." in post.author.publisher.isActive
func (query *QueryBuilder) addJoinFromFragments(scope *Scope, fragments []string) error {
	model := strcase.ToCamel(fragments[0])
	fragmentCount := len(fragments)

	for i := 1; i < fragmentCount-1; i++ {
		currentFragment := fragments[i]

		if !proto.ModelHasField(scope.schema, model, currentFragment) {
			return fmt.Errorf("this model: %s, does not have a field of name: %s", model, currentFragment)
		}

		// We know that the current fragment is a related model because it's not the last fragment
		relatedModelField := proto.FindField(scope.schema.Models, model, currentFragment)
		relatedModel := relatedModelField.Type.ModelName.Value
		identifierField := "id"

		switch {
		case proto.IsToOneRelationship(relatedModelField):
			foreignKeyField := relatedModelField.ForeignKeyFieldName.Value

			// Add a join to the primary key of the model that has-many in the M:1 relationship
			query.InnerJoin(relatedModel, ExpressionField(fragments[:i+1], identifierField), ExpressionField(fragments[:i], foreignKeyField))
		case proto.IsToManyRelationship(relatedModelField):
			fkModel := proto.FindModel(scope.schema.Models, relatedModelField.Type.ModelName.Value)
			fkField, found := lo.Find(fkModel.Fields, func(field *proto.Field) bool {
				return field.Type.Type == proto.Type_TYPE_MODEL && field.Type.ModelName.Value == model
			})
			if !found {
				return fmt.Errorf("no foreign key field found on related model %s", model)
			}

			foreignKeyField := fkField.ForeignKeyFieldName.Value

			// Add a join to the foreign key of the model that belongs-to in the 1:M relationship
			query.InnerJoin(relatedModel, ExpressionField(fragments[:i+1], foreignKeyField), ExpressionField(fragments[:i], identifierField))
		default:
			return fmt.Errorf("unhandled model relationship configuration for field: %s on model: %s", relatedModelField, relatedModelField.ModelName)
		}

		model = relatedModelField.Type.ModelName.Value
	}

	return nil
}

// Constructs a QueryOperand from a splice of fragments, representing an expression operand or implicit input.
// The fragment slice must include the base model as the first fragment, for example: post.author.publisher.isActive
func operandFromFragments(schema *proto.Schema, fragments []string) (*QueryOperand, error) {
	var field string
	model := strcase.ToCamel(fragments[0])
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
