package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
)

func (query *QueryBuilder) applyImplicitFiltersForList(scope *Scope, args map[string]any) error {
	message := proto.FindWhereInputMessage(scope.Schema, scope.Action.Name)
	if message == nil {
		return nil
	}

	return query.applyImplicitFiltersFromMessage(scope, message, scope.Model, args)
}

func (query *QueryBuilder) applyImplicitFiltersFromMessage(scope *Scope, message *proto.Message, model *proto.Model, args map[string]any) error {

	for _, input := range message.Fields {
		field := proto.FindField(scope.Schema.Models, model.Name, input.Name)

		// If the input is not targeting a model field, then it is either a:
		//  - Message, with nested fields which we must recurse into, or an
		//  - Explicit input, which is handled elsewhere.
		if !input.IsModelField() {
			if input.Type.Type == proto.Type_TYPE_MESSAGE {
				messageModel := proto.FindModel(scope.Schema.Models, field.Type.ModelName.Value)
				nestedMessage := proto.FindMessage(scope.Schema.Messages, input.Type.MessageName.Value)

				argsSectioned, ok := args[input.Name].(map[string]any)
				if !ok {
					if input.Optional {
						continue
					}
					return fmt.Errorf("cannot convert args to map[string]any for key: %s", input.Name)
				}

				err := query.applyImplicitFiltersFromMessage(scope, nestedMessage, messageModel, argsSectioned)
				if err != nil {
					return err
				}
			}
			continue
		}

		fieldName := input.Name
		value, ok := args[fieldName]

		// Not found in arguments
		if !ok {
			if input.Optional {
				continue
			}
			return fmt.Errorf("did not find required '%s' input in where clause", fieldName)
		}

		valueMap, ok := value.(map[string]any)

		// Cannot be parsed into map
		if !ok {
			return fmt.Errorf("'%s' input value %v is not in correct format", fieldName, value)
		}

		for operatorStr, operand := range valueMap {
			operator, err := graphQlOperatorToActionOperator(operatorStr)
			if err != nil {
				return err
			}

			// Resolve the database statement for this expression
			err = query.whereByImplicitFilter(scope, input.Target, fieldName, operator, operand)
			if err != nil {
				return err
			}

			// Implicit input conditions are ANDed together
			query.And()
		}
	}

	return nil
}

// Applies schema-defined @orderBy ordering to the query.
func (query *QueryBuilder) applySchemaOrdering(scope *Scope) error {
	for _, orderBy := range scope.Action.OrderBy {
		direction, err := toSql(orderBy.Direction)
		if err != nil {
			return err
		}

		query.AppendOrderBy(Field(orderBy.FieldName), direction)
	}

	return nil
}

// Applies ordering of @sortable fields to the query.
func (query *QueryBuilder) applyRequestOrdering(scope *Scope, orderBy []any) error {
	for _, item := range orderBy {
		obj := item.(map[string]any)
		for field, direction := range obj {
			query.AppendOrderBy(Field(field), direction.(string))
		}
	}

	return nil
}

func List(scope *Scope, input map[string]any) (map[string]any, error) {
	query := NewQuery(scope.Context, scope.Model)

	// Generate the SQL statement.
	statement, page, err := GenerateListStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	// Execute database request with results
	results, pageInfo, err := statement.ExecuteToMany(scope.Context, page)
	if err != nil {
		return nil, err
	}

	isAuthorised, err := AuthoriseAction(scope, input, results)
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, common.NewPermissionError()
	}

	return map[string]any{
		"results":  results,
		"pageInfo": pageInfo.ToMap(),
	}, nil
}

func GenerateListStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, *Page, error) {
	where, ok := input["where"].(map[string]any)
	if !ok {
		where = map[string]any{}
	}

	orderBy, ok := input["orderBy"].([]any)
	if !ok {
		orderBy = []any{}
	}

	err := query.applyImplicitFiltersForList(scope, where)
	if err != nil {
		return nil, nil, err
	}

	err = query.applyExplicitFilters(scope, where)
	if err != nil {
		return nil, nil, err
	}

	err = query.applySchemaOrdering(scope)
	if err != nil {
		return nil, nil, err
	}

	err = query.applyRequestOrdering(scope, orderBy)
	if err != nil {
		return nil, nil, err
	}

	page, err := ParsePage(input)
	if err != nil {
		return nil, nil, err
	}

	// Select all columns from this table and distinct on id
	query.AppendDistinctOn(IdField())
	query.AppendSelect(AllFields())

	err = query.ApplyPaging(page)
	if err != nil {
		return nil, &page, err
	}

	return query.SelectStatement(), &page, nil
}
