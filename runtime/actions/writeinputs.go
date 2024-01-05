package actions

import (
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/schema/parser"
)

// Updates the query with all set attributes defined on the action.
func (query *QueryBuilder) captureSetValues(scope *Scope, args map[string]any) error {
	model := proto.FindModel(scope.Schema.Models, strcase.ToCamel("identity"))
	ctxScope := NewModelScope(scope.Context, model, scope.Schema)
	identityQuery := NewQuery(model)

	identityId := ""
	if auth.IsAuthenticated(scope.Context) {
		identity, err := auth.GetIdentity(scope.Context)
		if err != nil {
			return err
		}
		identityId = identity.Id
	}

	err := identityQuery.Where(IdField(), Equals, Value(identityId))
	if err != nil {
		return err
	}

	for _, setExpression := range scope.Action.SetExpressions {
		expression, err := parser.ParseExpression(setExpression.Source)
		if err != nil {
			return err
		}

		assignment, err := expression.ToAssignmentCondition()
		if err != nil {
			return err
		}

		lhsResolver := expressions.NewOperandResolver(scope.Context, scope.Schema, scope.Model, scope.Action, assignment.LHS)
		rhsResolver := expressions.NewOperandResolver(scope.Context, scope.Schema, scope.Model, scope.Action, assignment.RHS)

		if !lhsResolver.IsModelDbColumn() {
			return errors.New("lhs operand of assignment expression must be a model field")
		}

		fragments, err := lhsResolver.NormalisedFragments()
		if err != nil {
			return err
		}

		currRows := []*Row{query.writeValues}

		// The model field to update.
		field := fragments[len(fragments)-1]
		targetsLessField := fragments[:len(fragments)-1]

		// If we are associating (as opposed to creating) then rather update the foreign key
		// i.e. person.employerId and not person.employer.id
		isAssoc := targetAssociating(scope, fragments)
		if isAssoc {
			field = fmt.Sprintf("%sId", fragments[len(fragments)-2])
			targetsLessField = fragments[:len(fragments)-2]
		}

		// Iterate through the fragments in the @set expression AND traverse the graph until we have a set of rows to update.
		for i, frag := range targetsLessField {
			nextRows := []*Row{}

			if len(currRows) == 0 {
				// We cannot set order.customer.name if the input is order.customer.id (implying an assocation).
				return fmt.Errorf("set expression operand out of range of inputs: %s. we currently only support setting fields within the input data and cannot set associated model fields", setExpression.Source)
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

		// Set the field on all rows.
		for _, row := range currRows {
			if rhsResolver.IsModelDbColumn() {

				rhsFragments, err := rhsResolver.NormalisedFragments()
				if err != nil {
					return err
				}

				row.values[field] = ExpressionField(rhsFragments, field)
			} else if rhsResolver.IsContextDbColumn() {
				// If this is a value from ctx that requires a database read (such as with identity backlinks),
				// then construct an inline query for this operand.  This is necessary because we can't retrieve this value
				// from the current query builder.

				fragments, err := rhsResolver.NormalisedFragments()
				if err != nil {
					return err
				}

				// Remove the ctx fragment
				fragments = fragments[1:]

				err = identityQuery.addJoinFromFragments(ctxScope, fragments)
				if err != nil {
					return err
				}

				selectField := ExpressionField(fragments[:len(fragments)-1], fragments[len(fragments)-1])

				identityQuery.AppendSelect(selectField)

				row.values[field] = InlineQuery(identityQuery, selectField)

			} else if rhsResolver.IsContextField() || rhsResolver.IsLiteral() || rhsResolver.IsExplicitInput() || rhsResolver.IsImplicitInput() {
				value, err := rhsResolver.ResolveValue(args)
				if err != nil {
					return err
				}

				row.values[field] = Value(value)
			}
		}
	}
	return nil
}

// Updates the query with all write inputs defined on the action.
func (query *QueryBuilder) captureWriteValues(scope *Scope, args map[string]any) error {
	message := proto.FindValuesInputMessage(scope.Schema, scope.Action.Name)
	if message == nil {
		return nil
	}

	target := []string{casing.ToLowerCamel(scope.Model.Name)}

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
	for _, input := range message.Fields {
		field := proto.FindField(scope.Schema.Models, model.Name, input.Name)

		// If the input is not targeting a model field, then it is either a:
		//  - Message, with nested fields which we must recurse into, or an
		//  - Explicit input, which is handled elsewhere.
		if !input.IsModelField() {
			if input.Type.Type == proto.Type_TYPE_MESSAGE {

				target := append(newRow.target, casing.ToLowerCamel(input.Name))
				messageModel := proto.FindModel(scope.Schema.Models, field.Type.ModelName.Value)
				nestedMessage := proto.FindMessage(scope.Schema.Messages, input.Type.MessageName.Value)

				var foreignKeys map[string]any
				var err error

				if proto.IsHasMany(field) || proto.IsHasOne(field) {
					// if proto.IsHasMany(field) then we have a 1:M relationship and the FK is on this model.
					// if proto.IsHasOne(field) then we have a 1:1 relationship with the FK on this model.

					arg, hasArg := args[input.Name]
					if !hasArg && !input.Optional {
						return nil, nil, fmt.Errorf("input argument is missing for required field %s", input.Name)
					} else if !hasArg && input.Optional {
						continue
					}

					var argsArraySectioned []any
					if proto.IsHasOne(field) {
						argsArraySectioned = []any{arg}
					} else {
						var ok bool
						argsArraySectioned, ok = arg.([]any)
						if !ok {
							return nil, nil, fmt.Errorf("cannot convert args to []any for key %s", input.Name)
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
						foriegnKeyModelField := lo.Filter(messageModel.Fields, func(f *proto.Field, i int) bool {
							return f.Type.Type == proto.Type_TYPE_MODEL &&
								f.Type.ModelName.Value == model.Name &&
								field != f && // For self-referencing models
								(field.InverseFieldName == nil || f.ForeignKeyFieldName.Value == fmt.Sprintf("%sId", field.InverseFieldName.Value))
						})

						if len(foriegnKeyModelField) != 1 {
							return nil, nil, fmt.Errorf("there needs to be exactly one foreign key field for %s", input.Name)
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
						return nil, nil, fmt.Errorf("it is not possible to specify just the id for %s as the foreign key does not sit on %s", input.Name, messageModel.Name)
					}
				} else {
					// A not-repeating field means that we have a M:1 or 1:1 relationship with the FK on the other model. Therefore:
					//  - we will have a single of model to parse,
					//  - this model will have the primary ID that needs to be referenced from the current model.

					argValue, hasArg := args[input.Name]
					if !hasArg {
						if !input.Optional {
							return nil, nil, fmt.Errorf("input argument is missing for required field %s", input.Name)
						}

						continue
					}

					if argValue == nil && !input.Nullable {
						return nil, nil, fmt.Errorf("input argument is null for non-nullable field %s", input.Name)
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
							return nil, nil, fmt.Errorf("cannot convert args to map[string]any for key %s", input.Name)
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
							foriegnKeyModelField := lo.Filter(model.Fields, func(f *proto.Field, _ int) bool {
								return f.Type.Type == proto.Type_TYPE_MODEL &&
									f.Type.ModelName.Value == messageModel.Name &&
									f.Name == input.Name
							})

							if len(foriegnKeyModelField) != 1 {
								return nil, nil, fmt.Errorf("there needs to be exactly one foreign key field for %s", input.Name)
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
			fieldName := fmt.Sprintf("%sId", input.Target[len(input.Target)-2])

			// Do not create a new row, and rather return this FK to add to the referencing row.
			return map[string]any{
				fieldName: args[input.Name],
			}, nil, nil
		} else {
			value, ok := args[input.Name]
			// Only add the arg value if it was provided as an input.
			if ok {
				newRow.values[input.Name] = Value(value)
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

	message := proto.FindMessage(scope.Schema.Messages, scope.Action.InputMessageName)
	model := proto.FindModel(scope.Schema.Models, strcase.ToCamel(target[0]))
	for _, t := range target[1 : len(target)-1] {
		found := false
		for _, f := range message.Fields {
			if f.Name == t {
				if f.Type.Type == proto.Type_TYPE_MESSAGE {
					message = proto.FindMessage(scope.Schema.Messages, f.Type.MessageName.Value)
					field := proto.FindField(scope.Schema.Models, model.Name, t)
					model = proto.FindModel(scope.Schema.Models, field.Type.ModelName.Value)

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
	for _, input := range message.Fields {
		// Skip named/non-model inputs
		if input.Target == nil {
			continue
		}

		field := proto.FindField(scope.Schema.Models, model.Name, input.Name)

		// If any non-id field is being set, then we are not associating.
		// This indicates that we are creating.
		if !field.PrimaryKey {
			return false
		}
	}
	return true
}
