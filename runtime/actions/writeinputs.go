package actions

import (
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/schema/parser"
)

// Updates the query with all set attributes defined on the operation.
func (query *QueryBuilder) captureSetValues(scope *Scope, args map[string]any) error {
	for _, setExpression := range scope.operation.SetExpressions {
		expression, err := parser.ParseExpression(setExpression.Source)
		if err != nil {
			return err
		}

		assignment, err := expression.ToAssignmentCondition()
		if err != nil {
			return err
		}

		lhsResolver := expressions.NewOperandResolver(scope.context, scope.schema, scope.operation, assignment.LHS)
		rhsResolver := expressions.NewOperandResolver(scope.context, scope.schema, scope.operation, assignment.RHS)
		operandType, err := lhsResolver.GetOperandType()
		if err != nil {
			return err
		}

		if !lhsResolver.IsDatabaseColumn() {
			return fmt.Errorf("lhs operand of assignment expression must be a model field")
		}

		value, err := rhsResolver.ResolveValue(args)
		if err != nil {
			return err
		}

		fieldName := assignment.LHS.Ident.Fragments[1].Fragment

		// Currently we only support 3 fragments in an set expression operand if it is targeting an "id" field.
		// If so, we generate the foreign key field name from the fragments.
		// For example, post.author.id will become authorId.
		if len(assignment.LHS.Ident.Fragments) == 3 {
			if assignment.LHS.Ident.Fragments[2].Fragment != "id" {
				return errors.New("currently only support 'id' as a third fragment in a set expression")
			}
			fieldName = fmt.Sprintf("%sId", fieldName)
		}

		// If targeting the nested model (without a field), then set the foreign key with the "id" of the assigning model.
		// For example, @set(post.user = ctx.identity) will set post.userId with ctx.identity.id.
		if operandType == proto.Type_TYPE_MODEL {
			fieldName = fmt.Sprintf("%sId", fieldName)
		}

		// Add a value to be written during an INSERT or UPDATE
		query.AddWriteValue(fieldName, value)

	}
	return nil
}

// Updates the query with all write inputs defined on the operation.
func (query *QueryBuilder) captureWriteValues(scope *Scope, args map[string]any) error {
	message := proto.FindValuesInputMessage(scope.schema, scope.operation.Name)
	if message == nil {
		return nil
	}

	foreignKeys, row, err := query.captureInputValuesFromMessage(scope, message, scope.model, args)

	for k, v := range foreignKeys {
		row.values[strcase.ToSnake(k)] = v
	}

	query.writeValues = row

	return err
}

func (query *QueryBuilder) captureInputValuesFromMessage(scope *Scope, message *proto.Message, model *proto.Model, args map[string]any) (map[string]any, *Row, error) {
	defaultValues := map[string]any{}
	var err error

	if scope.operation.Type == proto.OperationType_OPERATION_TYPE_CREATE {
		defaultValues, err = initialValueForModel(model, scope.schema)
		if err != nil {
			return nil, nil, err
		}
	}

	// Instantiate an empty row or with defaults if this is a create op
	newRow := &Row{
		model:        model,
		values:       defaultValues,
		referencedBy: []*Row{},
		references:   []*Row{},
	}

	// For each field in this model either:
	//   - add its values to the current row where an argument has been provided, OR
	//   - create a new row and relate it to the current row (either referencedBy or references), OR
	//   - determine that it is a primary key reference, do not create a row, and return the FK to the referencing row.
	for _, input := range message.Fields {
		field := proto.FindField(scope.schema.Models, model.Name, input.Name)

		// If the input is not targeting a model field, then it is either a:
		//  - Message, with nested fields which we must recurse into, or an
		//  - Explicit input, which is handled elsewhere.
		if !input.IsModelField() {
			if input.Type.Type == proto.Type_TYPE_MESSAGE {
				messageModel := proto.FindModel(scope.schema.Models, field.Type.ModelName.Value)
				nestedMessage := proto.FindMessage(scope.schema.Messages, input.Type.MessageName.Value)

				var foreignKeys map[string]any

				if input.Type.Repeated {
					argsArraySectioned, ok := args[input.Name].([]any)
					if !ok {
						return nil, nil, fmt.Errorf("cannot convert args to []any for key %s", input.Name)
					}

					var rows []*Row
					foreignKeys, rows, err = query.captureInputValuesArrayFromMessage(scope, nestedMessage, messageModel, argsArraySectioned)
					if err != nil {
						return nil, nil, err
					}

					if len(rows) > 0 {
						newRow.referencedBy = append(newRow.referencedBy, rows...)
					}

				} else {
					argsSectioned, ok := args[input.Name].(map[string]any)
					if !ok {
						return nil, nil, fmt.Errorf("cannot convert args to map[string]any for key %s", input.Name)
					}

					var row *Row
					foreignKeys, row, err = query.captureInputValuesFromMessage(scope, nestedMessage, messageModel, argsSectioned)
					if err != nil {
						return nil, nil, err
					}

					if row != nil {
						newRow.references = append(newRow.references, row)
					}
				}

				// If any nested messages referenced a primary key, then the
				// foreign keys will be generated instead of a new row created.
				for k, v := range foreignKeys {
					newRow.values[strcase.ToSnake(k)] = v
				}
			}

			continue
		}

		if field.PrimaryKey {
			// We know this needs to be a FK on the referencing row.
			fieldName := fmt.Sprintf("%sId", input.Target[len(input.Target)-2])

			// Do not create a new row, and rather return this FK to add to the referencing row.
			return map[string]any{
				fieldName: args[input.Name],
			}, nil, nil
		} else {
			value, ok := args[input.Name]
			// Only add the arg value if it was provided as an input.
			if ok {
				newRow.values[strcase.ToSnake(input.Name)] = value
			}
		}
	}

	return nil, newRow, nil
}

func (query *QueryBuilder) captureInputValuesArrayFromMessage(scope *Scope, message *proto.Message, model *proto.Model, argsArray []any) (map[string]any, []*Row, error) {
	rows := []*Row{}
	foreignKeys := map[string]any{}

	// Capture all fields for each item in the array.
	for _, v := range argsArray {
		args, ok := v.(map[string]any)
		if !ok {
			return nil, nil, errors.New("cannot convert args to map[string]any")
		}

		fks, row, err := query.captureInputValuesFromMessage(scope, message, model, args)
		if err != nil {
			return nil, nil, err
		}

		rows = append(rows, row)

		for k, v := range fks {
			foreignKeys[k] = v
		}
	}

	return foreignKeys, rows, nil
}
