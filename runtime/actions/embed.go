package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

func resolveEmbeddedData(ctx context.Context, schema *proto.Schema, sourceModel *proto.Model, sourceID string, fragments []string) (any, error) {
	if len(fragments) == 0 {
		return nil, errors.New("invalid embed resolver")
	}
	embedTargetField := fragments[0]

	field := proto.FindField(schema.Models, sourceModel.GetName(), casing.ToLowerCamel(embedTargetField))
	if field == nil {
		return nil, fmt.Errorf("embed target field (%s) does not exist in model %s", embedTargetField, sourceModel.GetName())
	}

	if !field.IsTypeModel() {
		return nil, fmt.Errorf("field (%s) is not a embeddable model field", embedTargetField)
	}

	// we will select from the relatedModel's table, joining the source model with an alias ("source"), where source.id = sourceModel.id
	relatedModelName := field.Type.ModelName.Value
	relatedModel := schema.FindModel(relatedModelName)
	foreignKeyField := proto.GetForeignKeyFieldName(schema.Models, field)
	sourceTableAlias := "_source"

	dbQuery := NewQuery(relatedModel)
	// we apply the where clause which will filter based on the joins set up depending on the relationship type
	err := dbQuery.Where(&QueryOperand{
		table:  sourceTableAlias,
		column: casing.ToSnake(parser.FieldNameId),
	}, Equals, Value(sourceID))
	if err != nil {
		return nil, fmt.Errorf("applying sql where: %w", err)
	}

	switch {
	case field.IsBelongsTo():
		dbQuery.Join(
			sourceModel.Name,
			&QueryOperand{
				table:  sourceTableAlias,
				column: casing.ToSnake(foreignKeyField),
			},
			&QueryOperand{
				table:  casing.ToSnake(relatedModelName),
				column: casing.ToSnake(parser.FieldNameId),
			})

		stmt := dbQuery.SelectStatement()
		result, err := stmt.ExecuteToSingle(ctx)
		if err != nil {
			return nil, fmt.Errorf("executing query to single: %w", err)
		}

		// recurse and resolve child embeds
		if len(fragments) > 1 {
			if childId, ok := result[parser.FieldNameId].(string); ok {
				childEmbed, err := resolveEmbeddedData(ctx, schema, relatedModel, childId, fragments[1:])
				if err != nil {
					return nil, fmt.Errorf("retrieving child embed: %w", err)
				}
				result[fragments[1]] = childEmbed

				// we now need to remove the foreign key field from the result (e.g. if we're embedding author, we want to remove authorId)
				delete(result, fragments[1]+"Id")
			}
		}

		// if we have any files in our results we need to transform them to the object structure required
		if relatedModel.HasFiles() {
			result, err = transformFileResponses(ctx, relatedModel, result)
			if err != nil {
				return nil, err
			}
		}

		return result, nil
	case field.IsHasMany():
		dbQuery.Join(
			sourceModel.Name,
			&QueryOperand{
				table:  sourceTableAlias,
				column: casing.ToSnake(parser.FieldNameId),
			},
			&QueryOperand{
				table:  casing.ToSnake(relatedModelName),
				column: casing.ToSnake(foreignKeyField),
			})
		stmt := dbQuery.SelectStatement()
		result, _, err := stmt.ExecuteToMany(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("executing query to many: %w", err)
		}

		// if we have any files in our results we need to go through each result and transform them to the object structure required
		if relatedModel.HasFiles() {
			for i := range result {
				result[i], err = transformFileResponses(ctx, relatedModel, result[i])
				if err != nil {
					return nil, err
				}
			}
		}

		// recurse and resolve child embeds for each of our results
		if len(fragments) > 1 {
			for i := range result {
				childId, ok := result[i][parser.FieldNameId].(string)
				if !ok {
					// we skip if we don't have a child embed id
					continue
				}
				childEmbed, err := resolveEmbeddedData(ctx, schema, relatedModel, childId, fragments[1:])
				if err != nil {
					return nil, fmt.Errorf("retrieving child embed: %w", err)
				}
				result[i][fragments[1]] = childEmbed

				// we now need to remove the foreign key field from the result (e.g. if we're embedding author, we want to remove authorId)
				delete(result[i], fragments[1]+"Id")
			}
		}

		return result, nil
	case field.IsHasOne():
		dbQuery.Join(
			sourceModel.Name,
			&QueryOperand{
				table:  sourceTableAlias,
				column: casing.ToSnake(parser.FieldNameId),
			},
			&QueryOperand{
				table:  casing.ToSnake(relatedModelName),
				column: casing.ToSnake(foreignKeyField),
			})

		stmt := dbQuery.SelectStatement()
		result, err := stmt.ExecuteToSingle(ctx)
		if err != nil {
			return nil, fmt.Errorf("executing query to single: %w", err)
		}

		if relatedModel.HasFiles() {
			// if we have any files in our results we need to transform them to the object structure required
			result, err = transformFileResponses(ctx, relatedModel, result)
			if err != nil {
				return nil, err
			}
		}

		// recurse and resolve child embeds
		if len(fragments) > 1 {
			if childId, ok := result[parser.FieldNameId].(string); ok {
				childEmbed, err := resolveEmbeddedData(ctx, schema, relatedModel, childId, fragments[1:])
				if err != nil {
					return nil, fmt.Errorf("retrieving child embed: %w", err)
				}
				result[fragments[1]] = childEmbed

				// we now need to remove the foreign key field from the result (e.g. if we're embedding author, we want to remove authorId)
				delete(result, fragments[1]+"Id")
			}
		}

		return result, nil
	}

	return nil, errors.New("unsupported embed type")
}
