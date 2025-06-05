package actions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/locale"
	"github.com/teamkeel/keel/schema/parser"
)

func (query *QueryBuilder) applyImplicitFiltersForList(scope *Scope, args map[string]any) error {
	message := proto.FindWhereInputMessage(scope.Schema, scope.Action.GetName())
	if message == nil {
		return nil
	}

	return query.applyImplicitFiltersFromMessage(scope, message, scope.Model, args)
}

func (query *QueryBuilder) applyImplicitFiltersFromMessage(scope *Scope, message *proto.Message, model *proto.Model, args map[string]any) error {
	for _, input := range message.GetFields() {
		field := proto.FindField(scope.Schema.GetModels(), model.GetName(), input.GetName())

		// If the input is not targeting a model field, then it is either a:
		//  - Message, with nested fields which we must recurse into, or an
		//  - Explicit input, which is handled elsewhere.
		if !input.IsModelField() {
			if input.GetType().GetType() == proto.Type_TYPE_MESSAGE {
				messageModel := scope.Schema.FindModel(field.GetType().GetModelName().GetValue())
				nestedMessage := scope.Schema.FindMessage(input.GetType().GetMessageName().GetValue())

				argsSectioned, ok := args[input.GetName()].(map[string]any)
				if !ok {
					if input.GetOptional() {
						continue
					}
					return fmt.Errorf("cannot convert args to map[string]any for key: %s", input.GetName())
				}

				err := query.applyImplicitFiltersFromMessage(scope, nestedMessage, messageModel, argsSectioned)
				if err != nil {
					return err
				}
			}
			continue
		}

		value, ok := args[input.GetName()]

		// Not found in arguments
		if !ok {
			if input.GetOptional() {
				continue
			}
			return fmt.Errorf("did not find required '%s' input in where clause", input.GetName())
		}

		valueMap, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("'%s' input value %v is not in correct format", input.GetName(), value)
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
					err = query.whereByImplicitFilter(scope, input.GetTarget(), operator, arrayOperand)
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
				err = query.whereByImplicitFilter(scope, input.GetTarget(), operator, operand)
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
	for _, orderBy := range scope.Action.GetOrderBy() {
		direction, err := toSql(orderBy.GetDirection())
		if err != nil {
			return err
		}

		query.AppendOrderBy(Field(orderBy.GetFieldName()), direction)
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
	canResolveEarly, authorised, err := TryResolveAuthorisationEarly(scope, input, permissions)
	if err != nil {
		return nil, err
	}
	if canResolveEarly && !authorised {
		return nil, common.NewPermissionError()
	}

	// Generate the SQL statement.
	opts := []QueryBuilderOption{}
	if location, err := locale.GetTimeLocation(scope.Context); err == nil {
		opts = append(opts, WithTimezone(location.String()))
	}
	query := NewQuery(scope.Model, opts...)

	statement, page, err := GenerateListStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	// Execute database request with results
	results, resultInfo, pageInfo, err := statement.ExecuteToMany(scope.Context, page)
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

	// if we have any files in our results we need to transform them to the object structure required
	if scope.Model.HasFiles() {
		for i := range results {
			results[i], err = transformModelFileResponses(scope.Context, scope.Model, results[i])
			if err != nil {
				return nil, fmt.Errorf("transforming file data: %w", err)
			}
		}
	}

	// if we have embedded data, let's resolve it
	if len(scope.Action.GetResponseEmbeds()) > 0 {
		for _, embed := range scope.Action.GetResponseEmbeds() {
			fragments := strings.Split(embed, ".")

			for _, res := range results {
				id, ok := res[parser.FieldNameId].(string)
				if !ok {
					return nil, errors.New("missing identifier")
				}
				data, err := resolveEmbeddedData(scope.Context, scope.Schema, scope.Model, id, fragments)
				if err != nil {
					return nil, err
				}
				res[fragments[0]] = data

				// we now need to remove the foreign key field from the result (e.g. if we're embedding author, we want to remove authorId)
				delete(res, fragments[0]+"Id")
			}
		}
	}

	res := map[string]any{
		"results":  results,
		"pageInfo": pageInfo.ToMap(),
	}

	if resultInfo != nil {
		res["resultInfo"] = resultInfo
	}

	return res, nil
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

	err := query.SelectFacets(scope, input)
	if err != nil {
		return nil, nil, err
	}

	err = query.applyImplicitFiltersForList(scope, where)
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
	query.DistinctOn(IdField())
	query.Select(AllFields())

	err = query.ApplyPaging(page)
	if err != nil {
		return nil, nil, err
	}

	return query.SelectStatement(), &page, nil
}
