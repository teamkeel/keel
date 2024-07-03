package actions

import (
	"errors"
	"fmt"
	"strings"

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

		value, ok := args[input.Name]

		// Not found in arguments
		if !ok {
			if input.Optional {
				continue
			}
			return fmt.Errorf("did not find required '%s' input in where clause", input.Name)
		}

		valueMap, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("'%s' input value %v is not in correct format", input.Name, value)
		}

		for operatorStr, operand := range valueMap {
			var operator ActionOperator
			var err error

			switch operatorStr {
			case "any", "all":
				arrayQueryValueMap, ok := value.(map[string]any)
				if !ok {
					return fmt.Errorf("'%s' input value %v is not in correct format", operatorStr, value)
				}

				for arrayOperatorStr, arrayOperand := range arrayQueryValueMap[operatorStr].(map[string]any) {
					switch operatorStr {
					case "any":
						operator, err = anyQueryOperationToActionOperator(arrayOperatorStr)
					case "all":
						operator, err = allQueryOperatorToActionOperator(arrayOperatorStr)
					}

					if err != nil {
						return err
					}

					// Resolve the database statement for this expression
					err = query.whereByImplicitFilter(scope, input.Target, operator, arrayOperand)
					if err != nil {
						return err
					}

					// Implicit input conditions are ANDed together
					query.And()
				}

			default:
				operator, err = queryOperatorToActionOperator(operatorStr)
				if err != nil {
					return err
				}

				// Resolve the database statement for this expression
				err = query.whereByImplicitFilter(scope, input.Target, operator, operand)
				if err != nil {
					return err
				}

				// Implicit input conditions are ANDed together
				query.And()
			}

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
func (query *QueryBuilder) applyRequestOrdering(orderBy []any) {
	for _, item := range orderBy {
		obj := item.(map[string]any)
		for field, direction := range obj {
			query.AppendOrderBy(Field(field), direction.(string))
		}
	}
}

func List(scope *Scope, input map[string]any) (map[string]any, error) {
	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)

	// Attempt to resolve permissions early; i.e. before row-based database querying.
	canResolveEarly, authorised, err := TryResolveAuthorisationEarly(scope, permissions)
	if err != nil {
		return nil, err
	}
	if canResolveEarly && !authorised {
		return nil, common.NewPermissionError()
	}

	// Generate the SQL statement.
	query := NewQuery(scope.Model)
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

	// if we have embedded data, let's resolve it
	if len(scope.Action.ResponseEmbeds) > 0 {
		for _, embed := range scope.Action.ResponseEmbeds {
			fragments := strings.Split(embed, ".")

			for _, res := range results {
				id, ok := res["id"].(string)
				if !ok {
					return nil, errors.New("missing identifier")
				}
				data, err := resolveEmbeddedData(scope.Context, scope.Schema, scope.Model, id, fragments)
				if err != nil {
					return nil, err
				}
				res[fragments[0]] = data
			}
		}
	}
	// if we have any files in our results we need to transform them to the object structure required
	for i := range results {
		results[i], err = transformFileResponses(scope, results[i])
		if err != nil {
			return nil, fmt.Errorf("transforming file data: %w", err)
		}
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

	err = query.applyExpressionFilters(scope, where)
	if err != nil {
		return nil, nil, err
	}

	err = query.applySchemaOrdering(scope)
	if err != nil {
		return nil, nil, err
	}

	query.applyRequestOrdering(orderBy)

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
