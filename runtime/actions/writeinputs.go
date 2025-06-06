package actions

import (
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// Updates the query with all set attributes defined on the action.
func (query *QueryBuilder) captureSetValues(scope *Scope, args map[string]any) error {
	for _, setExpression := range scope.Action.GetSetExpressions() {
		expression, err := parser.ParseExpression(setExpression.GetSource())
		if err != nil {
			return err
		}

		lhs, rhs, err := expression.ToAssignmentExpression()
		if err != nil {
			return err
		}

		target, err := resolve.AsIdent(lhs)
		if err != nil {
			return err
		}

		ident, err := NormalisedFragments(scope.Schema, target.Fragments)
		if err != nil {
			return err
		}

		currRows := []*Row{query.writeValues}

		// The model field to update.
		field := ident[len(ident)-1]
		targetsLessField := ident[:len(target.Fragments)-1]

		// If we are associating (as opposed to creating) then rather update the foreign key
		// i.e. person.employerId and not person.employer.id
		isAssoc := targetAssociating(scope, ident)
		if isAssoc {
			field = fmt.Sprintf("%sId", ident[len(ident)-2])
			targetsLessField = ident[:len(target.Fragments)-2]
		}

		// Iterate through the fragments in the @set expression AND traverse the graph until we have a set of rows to update.
		for i, frag := range targetsLessField {
			nextRows := []*Row{}

			if len(currRows) == 0 {
				// We cannot set order.customer.name if the input is order.customer.id (implying an association).
				return fmt.Errorf("set expression operand out of range of inputs: %s. we currently only support setting fields within the input data and cannot set associated model fields", setExpression.GetSource())
			}

			for _, row := range currRows {
				if frag == row.target[i] {
					if i == len(targetsLessField)-1 {
						nextRows = append(nextRows, row)
					} else {
						for _, ref := range row.references {
							nextRows = append(nextRows, ref.row)
						}
						for _, refBy := range row.referencedBy {
							nextRows = append(nextRows, refBy.row)
						}
					}
				}
			}

			currRows = nextRows
		}

		operand, err := resolve.RunCelVisitor(rhs, GenerateSelectQuery(scope.Context, scope.Schema, scope.Model, scope.Action, args))
		if err != nil {
			return err
		}

		// Set the field on all rows.
		for _, row := range currRows {
			row.values[field] = operand
		}
	}
	return nil
}

// Updates the query with all write inputs defined on the action.
func (query *QueryBuilder) captureWriteValues(scope *Scope, args map[string]any) error {
	target := []string{casing.ToLowerCamel(scope.Model.GetName())}

	message := proto.FindValuesInputMessage(scope.Schema, scope.Action.GetName())
	if message == nil {
		query.writeValues = &Row{
			model:        scope.Model,
			target:       target,
			values:       map[string]*QueryOperand{},
			referencedBy: []*Relationship{},
			references:   []*Relationship{},
		}
		return nil
	}

	foreignKeys, row, err := query.captureWriteValuesFromMessage(scope, message, scope.Model, target, args)

	// Add any foreign keys to the root row from rows which it references.
	for k, v := range foreignKeys {
		row.values[k] = Value(v)
	}

	query.writeValues = row

	return err
}

// Parses the input data and builds a graph of row data which is organised by how this data would be stored in the database.
// Uses the protobuf schema to determine which rows are referenced by using (i.e. it determines where the foreign key sits).
func (query *QueryBuilder) captureWriteValuesFromMessage(scope *Scope, message *proto.Message, model *proto.Model, currentTarget []string, args map[string]any) (map[string]any, *Row, error) {
	// Instantiate an empty row.
	newRow := &Row{
		model:        model,
		target:       currentTarget,
		values:       map[string]*QueryOperand{},
		referencedBy: []*Relationship{},
		references:   []*Relationship{},
	}

	// For each field in this message either:
	//   - add its value to the current row where an input has been provided, OR
	//   - create a new row and relate it to the current row (either referencedBy or references), OR
	//   - determine that it is a primary key reference, do not create a row, and return the FK to the referencing row.
	for _, input := range message.GetFields() {
		field := proto.FindField(scope.Schema.GetModels(), model.GetName(), input.GetName())

		// If the input is not targeting a model field, then it is either a:
		//  - Message, with nested fields which we must recurse into, or an
		//  - Explicit input, which is handled elsewhere.
		if !input.IsModelField() {
			if input.GetType().GetType() == proto.Type_TYPE_MESSAGE {
				target := append(newRow.target, casing.ToLowerCamel(input.GetName()))
				messageModel := scope.Schema.FindModel(field.GetType().GetModelName().GetValue())
				nestedMessage := scope.Schema.FindMessage(input.GetType().GetMessageName().GetValue())

				var foreignKeys map[string]any
				var err error

				if field.IsHasMany() || field.IsHasOne() {
					// if field.IsHasMany() then we have a 1:M relationship and the FK is on this model.
					// if field.IsHasOne() then we have a 1:1 relationship with the FK on this model.

					arg, hasArg := args[input.GetName()]
					if !hasArg && !input.GetOptional() {
						return nil, nil, fmt.Errorf("input argument is missing for required field %s", input.GetName())
					} else if !hasArg && input.GetOptional() {
						continue
					}

					var argsArraySectioned []any
					if field.IsHasOne() {
						argsArraySectioned = []any{arg}
					} else {
						var ok bool
						argsArraySectioned, ok = arg.([]any)
						if !ok {
							return nil, nil, fmt.Errorf("cannot convert args to []any for key %s", input.GetName())
						}
					}

					// Create (or associate with) all the models which this model will be referenced by.
					var rows []*Row
					foreignKeys, rows, err = query.captureWriteValuesArrayFromMessage(scope, nestedMessage, messageModel, target, argsArraySectioned)
					if err != nil {
						return nil, nil, err
					}

					// rows will be empty if we are associating to existing models.
					if len(rows) > 0 {
						// Retrieve the foreign key model field on the related model.
						// If there are multiple relationships to the same model, then field.InverseFieldName will be
						// populated and will provide the disambiguation as to which foreign key field to use.
						foriegnKeyModelField := lo.Filter(messageModel.GetFields(), func(f *proto.Field, i int) bool {
							return f.GetType().GetType() == proto.Type_TYPE_MODEL &&
								f.GetType().GetModelName().GetValue() == model.GetName() &&
								field != f && // For self-referencing models
								(field.GetInverseFieldName() == nil || f.GetForeignKeyFieldName().GetValue() == fmt.Sprintf("%sId", field.GetInverseFieldName().GetValue()))
						})

						if len(foriegnKeyModelField) != 1 {
							return nil, nil, fmt.Errorf("there needs to be exactly one foreign key field for %s", input.GetName())
						}

						for _, r := range rows {
							for _, fk := range foriegnKeyModelField {
								relationship := &Relationship{
									foreignKey: fk,
									row:        r,
								}
								newRow.referencedBy = append(newRow.referencedBy, relationship)
							}
						}
					} else if len(foreignKeys) > 0 {
						return nil, nil, fmt.Errorf("it is not possible to specify just the id for %s as the foreign key does not sit on %s", input.GetName(), messageModel.GetName())
					}
				} else {
					// A not-repeating field means that we have a M:1 or 1:1 relationship with the FK on the other model. Therefore:
					//  - we will have a single of model to parse,
					//  - this model will have the primary ID that needs to be referenced from the current model.

					argValue, hasArg := args[input.GetName()]
					if !hasArg {
						if !input.GetOptional() {
							return nil, nil, fmt.Errorf("input argument is missing for required field %s", input.GetName())
						}

						continue
					}

					if argValue == nil && !input.GetNullable() {
						return nil, nil, fmt.Errorf("input argument is null for non-nullable field %s", input.GetName())
					}

					if argValue == nil {
						// We know this needs to be a FK on the referencing row.
						fieldName := fmt.Sprintf("%sId", target[len(target)-1])
						foreignKeys = map[string]any{
							fieldName: nil,
						}
					} else {
						argsSectioned, ok := argValue.(map[string]any)
						if !ok {
							return nil, nil, fmt.Errorf("cannot convert args to map[string]any for key %s", input.GetName())
						}

						// Create (or associate with) the model which this model references.
						var row *Row
						foreignKeys, row, err = query.captureWriteValuesFromMessage(scope, nestedMessage, messageModel, target, argsSectioned)
						if err != nil {
							return nil, nil, err
						}

						// row will be nil if we are associating to an existing model.
						if row != nil {
							// Retrieve the foreign key model field on the this model.
							foriegnKeyModelField := lo.Filter(model.GetFields(), func(f *proto.Field, _ int) bool {
								return f.GetType().GetType() == proto.Type_TYPE_MODEL &&
									f.GetType().GetModelName().GetValue() == messageModel.GetName() &&
									f.GetName() == input.GetName()
							})

							if len(foriegnKeyModelField) != 1 {
								return nil, nil, fmt.Errorf("there needs to be exactly one foreign key field for %s", input.GetName())
							}

							// Add foreign key to current model from the newly referenced models.
							relationship := &Relationship{
								foreignKey: foriegnKeyModelField[0],
								row:        row,
							}
							newRow.references = append(newRow.references, relationship)
						}
					}
				}

				// If any nested messages referenced a primary key, then the
				// foreign keys will be generated instead of a new row created.
				for k, v := range foreignKeys {
					newRow.values[k] = Value(v)
				}
			}

			continue
		}

		if messageAssociating(scope, message, model) && len(currentTarget) > 1 {
			// We know this needs to be a FK on the referencing row.
			fieldName := fmt.Sprintf("%sId", input.GetTarget()[len(input.GetTarget())-2])

			// Do not create a new row, and rather return this FK to add to the referencing row.
			return map[string]any{
				fieldName: args[input.GetName()],
			}, nil, nil
		} else {
			value, ok := args[input.GetName()]
			// Only add the arg value if it was provided as an input.
			if ok {
				newRow.values[input.GetName()] = Value(value)
			}
		}
	}

	return nil, newRow, nil
}

func (query *QueryBuilder) captureWriteValuesArrayFromMessage(scope *Scope, message *proto.Message, model *proto.Model, currentTarget []string, argsArray []any) (map[string]any, []*Row, error) {
	rows := []*Row{}
	foreignKeys := map[string]any{}

	// Capture all fields for each item in the array.
	for _, v := range argsArray {
		args, ok := v.(map[string]any)
		if !ok {
			return nil, nil, errors.New("cannot convert args to map[string]any")
		}

		fks, row, err := query.captureWriteValuesFromMessage(scope, message, model, currentTarget, args)
		if err != nil {
			return nil, nil, err
		}

		if row != nil {
			rows = append(rows, row)
		}

		for k, v := range fks {
			foreignKeys[k] = v
		}
	}

	return foreignKeys, rows, nil
}

// Is this target level used to associate to an existing model,
// or are we creating a new row in the database?  This is determined
// by whether any non-id fields are being set.  Note that 'id' can be
// set in both _association_ and _creation_ forms.
func targetAssociating(scope *Scope, target []string) bool {
	// We are always creating the root model of the action
	if len(target) < 3 {
		return false
	}

	if scope.Action.GetInputMessageName() == "" {
		return true
	}

	message := scope.Schema.FindMessage(scope.Action.GetInputMessageName())
	model := scope.Schema.FindModel(strcase.ToCamel(target[0]))
	for _, t := range target[1 : len(target)-1] {
		found := false

		for _, f := range message.GetFields() {
			if f.GetName() == t {
				if f.GetType().GetType() == proto.Type_TYPE_MESSAGE {
					message = scope.Schema.FindMessage(f.GetType().GetMessageName().GetValue())
					field := proto.FindField(scope.Schema.GetModels(), model.GetName(), t)
					model = scope.Schema.FindModel(field.GetType().GetModelName().GetValue())
				}
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	return messageAssociating(scope, message, model)
}

// If this (nested) message going to be used to establish a related association, or
// are we creating the related data?
func messageAssociating(scope *Scope, message *proto.Message, model *proto.Model) bool {
	for _, input := range message.GetFields() {
		// Skip named/non-model inputs
		if input.Target == nil {
			continue
		}

		field := proto.FindField(scope.Schema.GetModels(), model.GetName(), input.GetName())

		// If any non-id field is being set, then we are not associating.
		// This indicates that we are creating.
		if !field.GetPrimaryKey() {
			return false
		}
	}
	return true
}
