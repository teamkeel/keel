package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/schema/parser"
)

// GenerateSelectQuery visits the expression and adds select clauses to the provided query builder.
func GenerateSelectQuery(ctx context.Context, schema *proto.Schema, model *proto.Model, action *proto.Action, inputs map[string]any) resolve.Visitor[*QueryOperand] {
	return &setQueryGen{
		ctx:    ctx,
		schema: schema,
		model:  model,
		action: action,
		inputs: inputs,
	}
}

var _ resolve.Visitor[*QueryOperand] = new(setQueryGen)

type setQueryGen struct {
	ctx     context.Context
	operand *QueryOperand
	schema  *proto.Schema
	model   *proto.Model
	action  *proto.Action
	inputs  map[string]any
}

func (v *setQueryGen) StartTerm(parenthesis bool) error {
	return nil
}

func (v *setQueryGen) EndTerm(parenthesis bool) error {
	return nil
}

func (v *setQueryGen) StartFunction(name string) error {
	return nil
}

func (v *setQueryGen) EndFunction() error {
	return nil
}

func (v *setQueryGen) StartArgument(num int) error {
	return nil
}

func (v *setQueryGen) EndArgument() error {
	return nil
}

func (v *setQueryGen) VisitAnd() error {
	return errors.New("and operator not supported with set")
}

func (v *setQueryGen) VisitOr() error {
	return errors.New("or operator not supported with set")
}

func (v *setQueryGen) VisitNot() error {
	return nil
}

func (v *setQueryGen) VisitOperator(op string) error {
	return fmt.Errorf("%s operator not supported with set", op)
}

func (v *setQueryGen) VisitLiteral(value any) error {
	if value == nil {
		v.operand = Null()
	} else {
		v.operand = Value(value)
	}
	return nil
}

func (v *setQueryGen) VisitIdent(ident *parser.ExpressionIdent) error {
	model := v.schema.FindModel(strcase.ToCamel(ident.Fragments[0]))

	normalised, err := NormaliseFragments(v.schema, ident.Fragments)
	if err != nil {
		return err
	}

	if model != nil && expressions.IsEntityDbColumn(model, normalised) && len(normalised) > 2 {
		switch v.action.GetType() {
		case proto.ActionType_ACTION_TYPE_CREATE:
			// This section performs a field lookup on a to-be related model as an inline query

			field := model.FindField(ident.Fragments[1])
			model = v.schema.FindModel(field.GetType().GetEntityName().GetValue())
			query := NewQuery(model)

			relatedModelField := v.model.FindField(normalised[1])

			subFragments := normalised[1:]
			subFragments[0] = strcase.ToLowerCamel(relatedModelField.GetType().GetEntityName().GetValue())

			err := query.AddJoinFromFragments(v.schema, subFragments)
			if err != nil {
				return err
			}

			id := v.inputs[subFragments[0]].(map[string]any)[parser.FieldNameId]
			err = query.Where(IdField(), Equals, Value(id))
			if err != nil {
				return err
			}

			selectField, err := operandFromFragments(v.schema, ident.Fragments[1:])
			if err != nil {
				return err
			}

			if selectField.IsArrayField() {
				query.SelectUnnested(selectField)
			} else {
				query.Select(selectField)
			}

			v.operand = InlineQuery(query, selectField)
		case proto.ActionType_ACTION_TYPE_UPDATE:
			// This section performs a field lookup on a related model as an inline query

			query := NewQuery(v.model)
			err = query.AddJoinFromFragments(v.schema, normalised)
			if err != nil {
				return err
			}

			// This takes into consideration unique and composite lookups
			for k, v := range v.inputs["where"].(map[string]any) {
				err = query.Where(Field(k), Equals, Value(v))
				if err != nil {
					return err
				}
			}

			selectField, err := operandFromFragments(v.schema, ident.Fragments)
			if err != nil {
				return err
			}

			if selectField.IsArrayField() {
				query.SelectUnnested(selectField)
			} else {
				query.Select(selectField)
			}

			v.operand = InlineQuery(query, selectField)
		}
	} else {
		if v.action.GetType() == proto.ActionType_ACTION_TYPE_UPDATE && v.inputs["values"] != nil {
			v.inputs = v.inputs["values"].(map[string]any)
		}

		operand, err := generateOperand(v.ctx, v.schema, v.model, v.action, v.inputs, ident.Fragments)
		if err != nil {
			return err
		}
		v.operand = operand
	}

	return nil
}

func (v *setQueryGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	arr := []string{}
	for _, e := range idents {
		arr = append(arr, e.Fragments[1])
	}

	v.operand = Value(arr)

	return nil
}

func (v *setQueryGen) Result() (*QueryOperand, error) {
	return v.operand, nil
}
