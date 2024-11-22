package actions

import (
	"errors"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/schema/parser"
)

func Get(scope *Scope, input map[string]any) (map[string]any, error) {
	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)

	// Attempt to resolve permissions early; i.e. before row-based database querying.
	canResolveEarly, authorised, err := TryResolveAuthorisationEarly(scope, permissions)
	if err != nil {
		return nil, err
	}
	if canResolveEarly && !authorised {
		return nil, common.NewPermissionError()
	}

	// Generate the SQL statement
	query := NewQuery(scope.Model)
	statement, err := GenerateGetStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	// Execute database request, expecting a single result.
	res, err := statement.ExecuteToSingle(scope.Context)
	if err != nil {
		return nil, err
	}

	rowsToAuthorise := []map[string]any{}
	if res != nil {
		rowsToAuthorise = append(rowsToAuthorise, res)
	}

	isAuthorised, err := AuthoriseAction(scope, input, rowsToAuthorise)
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, common.NewPermissionError()
	}

	// if we have any files in our results we need to transform them to the object structure required
	if scope.Model.HasFiles() {
		res, err = transformModelFileResponses(scope.Context, scope.Model, res)
		if err != nil {
			return nil, err
		}
	}

	// if we have embedded data, let's resolve it
	if len(scope.Action.ResponseEmbeds) > 0 {
		for _, embed := range scope.Action.ResponseEmbeds {
			// a response embed will be something like: `book.author.country` where book is the model for the current action
			fragments := strings.Split(embed, ".")

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

	return res, err
}

func GenerateGetStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, error) {
	err := query.ApplyImplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	// err = query.applyExpressionFilters(scope, input)
	// if err != nil {
	// 	return nil, err
	// }

	err = query.applyExpressionFilters(scope, input)
	if err != nil {
		return nil, err
	}

	// Select all columns and distinct on id
	query.Select(AllFields())
	query.DistinctOn(IdField())

	return query.SelectStatement(), nil
}
