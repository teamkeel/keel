package actions

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/actions")

const (
	authenticateOperationName         = "authenticate"
	requestPasswordResetOperationName = "requestPasswordReset"
	passwordResetOperationName        = "resetPassword"
)

type Scope struct {
	Context   context.Context
	Operation *proto.Operation
	Model     *proto.Model
	Schema    *proto.Schema
}

func (s *Scope) WithContext(ctx context.Context) *Scope {
	return &Scope{
		Context:   ctx,
		Operation: s.Operation,
		Model:     s.Model,
		Schema:    s.Schema,
	}
}

func NewScope(
	ctx context.Context,
	operation *proto.Operation,
	schema *proto.Schema) *Scope {

	model := proto.FindModel(schema.Models, operation.ModelName)

	return &Scope{
		Context:   ctx,
		Operation: operation,
		Model:     model,
		Schema:    schema,
	}
}

func NewModelScope(
	ctx context.Context,
	model *proto.Model,
	schema *proto.Schema) *Scope {

	return &Scope{
		Context:   ctx,
		Operation: nil,
		Model:     model,
		Schema:    schema,
	}
}

func Execute(scope *Scope, inputs any) (any, map[string][]string, error) {
	ctx, span := tracer.Start(scope.Context, scope.Operation.Name)
	defer span.End()

	span.SetAttributes(
		attribute.String("action", scope.Operation.Name),
		attribute.String("model", scope.Model.Name),
	)

	scope = scope.WithContext(ctx)

	// inputs can be anything - with arbitrary functions 'Any' type, they can be
	// an array / number / string etc, which doesn't fit in with the traditional map[string]any definition of an inputs object
	inputsAsMap, inputWasAMap := inputs.(map[string]any)

	switch scope.Operation.Implementation {
	case proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM:
		return executeCustomFunction(scope, inputs)
	case proto.OperationImplementation_OPERATION_IMPLEMENTATION_RUNTIME:
		if !inputWasAMap {
			if inputs == nil {
				inputsAsMap = make(map[string]any)
			} else {
				return nil, nil, fmt.Errorf("inputs %v were not in correct format", inputs)
			}
		}
		return executeRuntimeOperation(scope, inputsAsMap)
	case proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO:
		if !inputWasAMap {
			if inputs == nil {
				inputsAsMap = make(map[string]any)
			} else {
				return nil, nil, fmt.Errorf("inputs %v were not in correct format", inputs)
			}
		}
		return executeAutoOperation(scope, inputsAsMap)
	default:
		return nil, nil, fmt.Errorf("unhandled unknown operation %s of type %s", scope.Operation.Name, scope.Operation.Implementation)
	}
}

func executeCustomFunction(scope *Scope, inputs any) (any, map[string][]string, error) {
	rolePermissions := proto.PermissionsForAction(scope.Schema, scope.Operation, func(p *proto.PermissionRule) bool {
		return len(p.RoleNames) > 0
	})

	permissionState := common.NewPermissionState()

	if runtimectx.IsAuthenticated(scope.Context) {
		granted, err := roleBasedPermissionGranted(scope, rolePermissions)

		if err != nil {
			return nil, nil, err
		}

		if granted {
			permissionState.Grant(common.GrantReasonRole)
		}
	}

	// todo: evaluate static expressions that don't need to hit the database
	// and grant them with reason "GrantReasonExpression"

	resp, headers, err := functions.CallFunction(
		scope.Context,
		scope.Operation.Name,
		inputs,
		permissionState,
	)

	if err != nil {
		return nil, nil, err
	}

	// For now a custom list function just returns a list of records, but the API's
	// all return an objects containing results and pagination info. So we need
	// to "wrap" the results here.
	// TODO: come up with a better implementation for list functions that can support
	// pagination
	if scope.Operation.Type == proto.OperationType_OPERATION_TYPE_LIST {
		results, _ := resp.([]any)
		return map[string]any{
			"results": results,
			"pageInfo": map[string]any{
				// todo: need to get these values from custom function return value
				// once we have changed the return type in the codegen and made changes
				// to the model api to support paging in some guise.
				"hasNextPage": false,
				"totalCount":  0,
				"count":       0,
				"startCursor": "",
				"endCursor":   "",
			},
		}, headers, nil
	}

	return resp, headers, err
}

func executeRuntimeOperation(scope *Scope, inputs map[string]any) (any, map[string][]string, error) {
	switch scope.Operation.Name {
	case authenticateOperationName:
		result, err := Authenticate(scope, inputs)
		return result, nil, err
	case requestPasswordResetOperationName:
		err := ResetRequestPassword(scope, inputs)
		return map[string]any{}, nil, err
	case passwordResetOperationName:
		err := ResetPassword(scope, inputs)
		return map[string]any{}, nil, err
	default:
		return nil, nil, fmt.Errorf("unhandled runtime operation: %s", scope.Operation.Name)
	}
}

func executeAutoOperation(scope *Scope, inputs map[string]any) (any, map[string][]string, error) {
	switch scope.Operation.Type {
	case proto.OperationType_OPERATION_TYPE_GET:
		v, err := Get(scope, inputs)
		// Get() can return nil, but for some reason if we don't explicitly
		// return nil here too the result becomes an empty map, which is rather
		// odd.
		// Simple repo of this: https://play.golang.com/p/MbBzvhrdOm_f
		if v == nil {
			return nil, nil, err
		}
		return v, nil, err
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		result, err := Update(scope, inputs)
		return result, nil, err
	case proto.OperationType_OPERATION_TYPE_CREATE:
		result, err := Create(scope, inputs)
		return result, nil, err
	case proto.OperationType_OPERATION_TYPE_DELETE:
		result, err := Delete(scope, inputs)
		return result, nil, err
	case proto.OperationType_OPERATION_TYPE_LIST:
		result, err := List(scope, inputs)
		return result, nil, err
	default:
		return nil, nil, fmt.Errorf("unhandled auto operation type: %s", scope.Operation.Type.String())
	}
}
