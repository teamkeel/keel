package actions

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/actions")

const (
	authenticateActionName         = "authenticate"
	requestPasswordResetActionName = "requestPasswordReset"
	passwordResetActionName        = "resetPassword"
)

type Scope struct {
	Context context.Context
	Action  *proto.Action
	Model   *proto.Model
	Job     *proto.Job
	Schema  *proto.Schema
}

func (s *Scope) WithContext(ctx context.Context) *Scope {
	return &Scope{
		Context: ctx,
		Action:  s.Action,
		Model:   s.Model,
		Schema:  s.Schema,
	}
}

func NewScope(
	ctx context.Context,
	action *proto.Action,
	schema *proto.Schema) *Scope {

	model := proto.FindModel(schema.Models, action.ModelName)

	return &Scope{
		Context: ctx,
		Action:  action,
		Model:   model,
		Job:     nil,
		Schema:  schema,
	}
}

func NewModelScope(
	ctx context.Context,
	model *proto.Model,
	schema *proto.Schema) *Scope {

	return &Scope{
		Context: ctx,
		Action:  nil,
		Model:   model,
		Job:     nil,
		Schema:  schema,
	}
}

func NewJobScope(
	ctx context.Context,
	job *proto.Job,
	schema *proto.Schema) *Scope {

	return &Scope{
		Context: ctx,
		Action:  nil,
		Model:   nil,
		Job:     job,
		Schema:  schema,
	}
}

func Execute(scope *Scope, inputs any) (result any, headers map[string][]string, err error) {
	ctx, span := tracer.Start(scope.Context, scope.Action.Name)
	defer span.End()

	span.SetAttributes(
		attribute.String("action", scope.Action.Name),
		attribute.String("model", scope.Model.Name),
	)

	scope = scope.WithContext(ctx)

	// Capture some request-oriented data and register it in the database config (briefly) for
	// the database audit function to pick up.
	if err := SetAuditScopeIntoDB(scope, span); err != nil {
		return nil, nil, err
	}
	defer ClearAuditScopeInDB(scope)

	// inputs can be anything - with arbitrary functions 'Any' type, they can be
	// an array / number / string etc, which doesn't fit in with the traditional map[string]any definition of an inputs object
	inputsAsMap, inputWasAMap := inputs.(map[string]any)

	switch scope.Action.Implementation {
	case proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM:
		result, headers, err = executeCustomFunction(scope, inputs)
	case proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME:
		if !inputWasAMap {
			if inputs == nil {
				inputsAsMap = make(map[string]any)
			} else {
				return nil, nil, fmt.Errorf("inputs %v were not in correct format", inputs)
			}
		}
		result, headers, err = executeRuntimeAction(scope, inputsAsMap)
	case proto.ActionImplementation_ACTION_IMPLEMENTATION_AUTO:
		if !inputWasAMap {
			if inputs == nil {
				inputsAsMap = make(map[string]any)
			} else {
				return nil, nil, fmt.Errorf("inputs %v were not in correct format", inputs)
			}
		}
		result, headers, err = executeAutoAction(scope, inputsAsMap)
	default:
		return nil, nil, fmt.Errorf("unhandled unknown action %s of type %s", scope.Action.Name, scope.Action.Implementation)
	}

	// Generate and send any events for this context.
	// If event sending fails, then record this in the span,
	// but do not return the error. The action should still succeed.
	eventsErr := events.SendEvents(ctx)
	if eventsErr != nil {
		span.RecordError(eventsErr)
	}

	return
}

func executeCustomFunction(scope *Scope, inputs any) (any, map[string][]string, error) {
	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)

	canAuthoriseEarly, authorised, err := TryResolveAuthorisationEarly(scope, permissions)
	if err != nil {
		return nil, nil, err
	}

	permissionState := common.NewPermissionState()
	if canAuthoriseEarly {
		if authorised {
			permissionState.Grant()
		} else {
			return nil, nil, common.NewPermissionError()
		}
	}

	resp, headers, err := functions.CallFunction(
		scope.Context,
		scope.Action.Name,
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
	if scope.Action.Type == proto.ActionType_ACTION_TYPE_LIST {
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

func executeRuntimeAction(scope *Scope, inputs map[string]any) (any, map[string][]string, error) {
	switch scope.Action.Name {
	case authenticateActionName:
		result, err := Authenticate(scope, inputs)
		return result, nil, err
	case requestPasswordResetActionName:
		err := ResetRequestPassword(scope, inputs)
		return map[string]any{}, nil, err
	case passwordResetActionName:
		err := ResetPassword(scope, inputs)
		return map[string]any{}, nil, err
	default:
		return nil, nil, fmt.Errorf("unhandled runtime action: %s", scope.Action.Name)
	}
}

func executeAutoAction(scope *Scope, inputs map[string]any) (any, map[string][]string, error) {
	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)

	// Attempt to resolve permissions early; i.e. before row-based database querying.
	canResolveEarly, authorised, err := TryResolveAuthorisationEarly(scope, permissions)
	if err != nil {
		return nil, nil, err
	}
	if canResolveEarly && !authorised {
		return nil, nil, common.NewPermissionError()
	}

	switch scope.Action.Type {
	case proto.ActionType_ACTION_TYPE_GET:
		v, err := Get(scope, inputs)
		// Get() can return nil, but for some reason if we don't explicitly
		// return nil here too the result becomes an empty map, which is rather
		// odd.
		// Simple repo of this: https://play.golang.com/p/MbBzvhrdOm_f
		if v == nil {
			return nil, nil, err
		}
		return v, nil, err
	case proto.ActionType_ACTION_TYPE_UPDATE:
		result, err := Update(scope, inputs)
		return result, nil, err
	case proto.ActionType_ACTION_TYPE_CREATE:
		result, err := Create(scope, inputs)
		return result, nil, err
	case proto.ActionType_ACTION_TYPE_DELETE:
		result, err := Delete(scope, inputs)
		return result, nil, err
	case proto.ActionType_ACTION_TYPE_LIST:
		result, err := List(scope, inputs)
		return result, nil, err
	default:
		return nil, nil, fmt.Errorf("unhandled auto action type: %s", scope.Action.Type.String())
	}
}
