package actions

import (
	"fmt"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
)

// Constructs and adds an LEFT JOIN from a splice of fragments (representing an operand in an expression or implicit input).
// The fragment slice must include the base model as the first item, for example: "post." in post.author.publisher.isActive
func (query *QueryBuilder) AddJoinFromFragments(schema *proto.Schema, fragments []string) error {
	model := casing.ToCamel(fragments[0])
	fragmentCount := len(fragments)

	for i := 1; i < fragmentCount-1; i++ {
		currentFragment := fragments[i]

		if !proto.ModelHasField(schema, model, currentFragment) {
			return fmt.Errorf("this model: %s, does not have a field of name: %s", model, currentFragment)
		}

		// We know that the current fragment is a related model because it's not the last fragment
		relatedModelField := proto.FindField(schema.Models, model, currentFragment)
		relatedModel := relatedModelField.Type.ModelName.Value
		foreignKeyField := proto.GetForeignKeyFieldName(schema.Models, relatedModelField)
		primaryKey := "id"

		var leftOperand *QueryOperand
		var rightOperand *QueryOperand

		switch {
		case relatedModelField.IsBelongsTo():
			// In a "belongs to" the foreign key is on _this_ model
			leftOperand = ExpressionField(fragments[:i+1], primaryKey, false)
			rightOperand = ExpressionField(fragments[:i], foreignKeyField, false)
		default:
			// In all others the foreign key is on the _other_ model
			leftOperand = ExpressionField(fragments[:i+1], foreignKeyField, false)
			rightOperand = ExpressionField(fragments[:i], primaryKey, false)
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
	isArray := false

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
			isArray = proto.FindField(schema.Models, model, currentFragment).Type.Repeated

		}
	}

	return ExpressionField(fragments[:len(fragments)-1], field, isArray), nil
}

// // Generates a database QueryOperand, either representing a field, inline query, a value or null.
// func generateQueryOperand(resolver *expressions.OperandResolver, args map[string]any) (*QueryOperand, error) {
// 	var queryOperand *QueryOperand

// 	switch {
// 	case resolver.IsContextDbColumn():
// 		// If this is a value from ctx that requires a database read (such as with identity backlinks),
// 		// then construct an inline query for this operand.  This is necessary because we can't retrieve this value
// 		// from the current query builder.

// 		fragments, err := resolver.NormalisedFragments()
// 		if err != nil {
// 			return nil, err
// 		}

// 		// Remove the ctx fragment
// 		fragments = fragments[1:]

// 		identityModel := resolver.Schema.FindModel(strcase.ToCamel(fragments[0]))
// 		ctxScope := NewModelScope(resolver.Context, identityModel, resolver.Schema)
// 		query := NewQuery(identityModel)

// 		identityId := ""
// 		if auth.IsAuthenticated(resolver.Context) {
// 			identity, err := auth.GetIdentity(resolver.Context)
// 			if err != nil {
// 				return nil, err
// 			}
// 			identityId = identity[parser.FieldNameId].(string)
// 		}

// 		err = query.addJoinFromFragments(ctxScope, fragments)
// 		if err != nil {
// 			return nil, err
// 		}

// 		err = query.Where(IdField(), Equals, Value(identityId))
// 		if err != nil {
// 			return nil, err
// 		}

// 		selectField := ExpressionField(fragments[:len(fragments)-1], fragments[len(fragments)-1])

// 		// If there are no matches in the subquery then null will be returned, but null
// 		// will cause IN and NOT IN filtering of this subquery result to always evaluate as false.
// 		// Therefore we need to filter out null.
// 		query.And()
// 		err = query.Where(selectField, NotEquals, Null())
// 		if err != nil {
// 			return nil, err
// 		}

// 		currModel := identityModel
// 		for i := 1; i < len(fragments)-1; i++ {
// 			name := proto.FindField(resolver.Schema.Models, currModel.Name, fragments[i]).Type.ModelName.Value
// 			currModel = resolver.Schema.FindModel(name)
// 		}
// 		currField := proto.FindField(resolver.Schema.Models, currModel.Name, fragments[len(fragments)-1])

// 		if currField.Type.Repeated {
// 			query.SelectUnnested(selectField)
// 		} else {
// 			query.Select(selectField)
// 		}

// 		queryOperand = InlineQuery(query, selectField)

// 	case resolver.IsModelDbColumn():
// 		// If this is a model field then generate the appropriate column operand for the database query.

// 		fragments, err := resolver.NormalisedFragments()
// 		if err != nil {
// 			return nil, err
// 		}

// 		// Generate QueryOperand from the fragments that make up the expression operand
// 		queryOperand, err = operandFromFragments(resolver.Schema, fragments)
// 		if err != nil {
// 			return nil, err
// 		}
// 	default:
// 		// For all others operands, we know we can resolve their value without the datebase.

// 		value, err := resolver.ResolveValue(args)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if value == nil {
// 			queryOperand = Null()
// 		} else {
// 			queryOperand = Value(value)
// 		}
// 	}

// 	return queryOperand, nil
// }
