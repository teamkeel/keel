package actions

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"golang.org/x/exp/maps"
)

type RootAction struct {
	inputs    map[string]any
	schema    *proto.Schema
	operation *proto.Operation
	ctx       context.Context

	// implicit and explicit inputs can clash
	// so resolve them here
	// explicit inputs always take precedence over implicitly set ones
	// in the final model map that is passed to the action
	reconciledInputs map[string]any
}

func (a *RootAction) WithSchema(schema *proto.Schema) *RootAction {
	a.schema = schema

	return a
}

func (a *RootAction) WithContext(ctx context.Context) *RootAction {
	a.ctx = ctx

	return a
}

func (a *RootAction) WithOperation(operation *proto.Operation) *RootAction {
	a.operation = operation

	return a
}

func (a *RootAction) Execute() (map[string]any, error) {
	if a.inputs == nil {
		panic("no args specified")
	}

	if a.operation == nil {
		panic("no operation specified")
	}

	switch a.operation.Type {
	case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
		// args:
		// { "" }
	case proto.OperationType_OPERATION_TYPE_DELETE:
		// handle args in format : { "k": "v" }
	case proto.OperationType_OPERATION_TYPE_LIST:
		// handle args in format : { "where": {}, "pageInfo": {}}
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		// handle args in format: { "where": {}, "values": {}}
	case proto.OperationType_OPERATION_TYPE_CREATE:
		// 1. build explicit inputs from where they are used in @set expressions
		// 2. build implicit inputs
		resolvedInputs, err := a.reconcileWriteInputs()

		if err != nil {
			return nil, err
		}

		result, err := Create(a.ctx, a.operation, a.schema, resolvedInputs)

		if err != nil {
			return nil, err
		}

		return result, nil
	}

	panic("not a known operation type")
}

// inputs - map[string]any - the input args for the operation
func (a *RootAction) WithInputs(inputs map[string]any) *RootAction {
	a.inputs = inputs

	return a
}

func (a *RootAction) model() *proto.Model {
	return proto.FindModel(a.schema.Models, a.operation.ModelName)
}

func (a *RootAction) reconcileReadInputs() (map[string]any, error) {
	queryMap := map[string]any{}

	// first we want to assign any implicit read inputs to the queryMap
	for _, input := range a.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		identifier := input.Target[0]

		valueFromArg, ok := a.inputs[identifier]

		if !input.Optional && !ok {
			return nil, fmt.Errorf("expected input %s missing from arguments %+v", identifier, a.inputs)
		}

		dbValue, err := toMap(valueFromArg, input.Type.Type)

		if err != nil {
			return nil, fmt.Errorf("could not convert %s value to database value", identifier)
		}

		queryMap[identifier] = dbValue
	}

	// then we want to overwrite any of the values set by implicit inputs with explicit ones
	// if they assign to the same fields, and assign any new values not set by implicit

	for _, where := range a.operation.WhereExpressions {

	}

	return queryMap, nil
}

func (a *RootAction) reconcileWriteInputs() (map[string]any, error) {

	// Complex: oneOf etc (list)

	// Simple = equals only + unique fields only (get, update, delete)

	modelMap := map[string]any{}

	// first assign implicit inputs and their values
	for _, input := range a.operation.Inputs {
		switch input.Behaviour {
		case proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT:
			modelFieldName := input.Target[0]

			v, matchingArg := a.inputs[input.Name]
			if input.Optional && !matchingArg {
				continue
			}

			if !matchingArg {
				return nil, fmt.Errorf("missing required input %s", input.Name)
			}

			v, err := toMap(v, input.Type.Type)

			if err != nil {
				return nil, err
			}

			modelMap[strcase.ToSnake(modelFieldName)] = v
		case proto.InputBehaviour_INPUT_BEHAVIOUR_EXPLICIT:
			continue
		default:
			return nil, fmt.Errorf("input behaviour %s is not yet supported for Create", input.Behaviour)
		}
	}

	setArgs, err := SetExpressionInputsToModelMap(a.operation, a.inputs, a.schema, a.ctx)

	if err != nil {
		return nil, err
	}

	// todo: clashing keys between implicit / explicit args (is this possible?)
	maps.Copy(modelMap, setArgs)

	maps.DeleteFunc(modelMap, func(k string, v any) bool {
		match := lo.SomeBy(a.model().Fields, func(f *proto.Field) bool {
			return strcase.ToSnake(f.Name) == k
		})

		return !match
	})

	return modelMap, nil
}

func (a *RootAction) convertToDatabaseMap(inputObject map[string]any) map[string]any {
	// we may not be able to use this until we have extracted out  addditional stuff

	panic("err")
}

func (a *RootAction) convertFromDbResult(result interface{}) interface{} {
	dbObject, isObject := result.(map[string]any)

	if isObject {
		return toLowerCamelMap(dbObject)
	}

	dbArray, isArray := result.([]map[string]any)

	if isArray {
		return toLowerCamelMaps(dbArray)
	}

	panic("not a valid db result")
}

// func main() {
// 	action := RootAction{}

// 	action.
// 		WithSchema(schema).
// 		WithContext(ctx).
// 		WithArgs(map[string]any{
// 			"test": "test",
// 		}).
// 		WithOperation(operation).
// 		Execute()
// }
