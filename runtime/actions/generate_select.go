package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// GenerateSelectQuery visits the expression and adds select clauses to the provided query builder
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
	field := proto.FindField(v.schema.Models, model.Name, ident.Fragments[1])

	normalised, err := NormalisedFragments(v.schema, ident.Fragments)
	if err != nil {
		return err
	}

	if len(normalised) == 3 {
		operand, err := generateOperand(v.ctx, v.schema, v.model, v.action, v.inputs, ident.Fragments)
		if err != nil {
			return err
		}
		v.operand = operand
	} else {
		// Join together all the tables based on the ident fragments
		model = v.schema.FindModel(field.Type.ModelName.Value)
		query := NewQuery(model)

		relatedModelField := proto.FindField(v.schema.Models, v.model.Name, normalised[1])
		subFragments := normalised[1:]
		subFragments[0] = strcase.ToLowerCamel(relatedModelField.Type.ModelName.Value)

		err := query.AddJoinFromFragments(v.schema, subFragments)
		if err != nil {
			return err
		}

		// Select the column as specified in the last ident fragment
		//fieldName := normalised[len(normalised)-1]
		//fragments := subFragments[:len(subFragments)-1]
		//query.Select(ExpressionField(fragments, fieldName, false))

		// Filter by this model's row's ID
		//foreignKeyField := proto.GetForeignKeyFieldName(v.schema.Models, relatedModelField)

		id := v.inputs[subFragments[0]].(map[string]any)["id"]

		//fk := fmt.Sprintf("r.\"%s\"", strcase.ToSnake(foreignKeyField))
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
