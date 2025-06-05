package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/actions")

const (
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
	model := schema.FindModel(action.GetModelName())

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

func Execute(scope *Scope, input any) (result any, meta *common.ResponseMetadata, err error) {
	ctx, span := tracer.Start(scope.Context, scope.Action.GetName())
	defer span.End()

	span.SetAttributes(
		attribute.String("action", scope.Action.GetName()),
		attribute.String("model", scope.Model.GetName()),
	)

	scope = scope.WithContext(ctx)

	// inputs can be anything - with arbitrary functions 'Any' type, they can be
	// an array / number / string etc, which doesn't fit in with the traditional map[string]any definition of an inputs object
	inputsAsMap, isMap := input.(map[string]any)

	if isMap && scope.Action.GetInputMessageName() != "" {
		message := scope.Schema.FindMessage(scope.Action.GetInputMessageName())
		isFunction := scope.Action.GetImplementation() == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM

		inputsAsMap, err = TransformInputs(scope.Schema, message, inputsAsMap, isFunction)
		if err != nil {
			return nil, nil, err
		}
	}

	switch scope.Action.GetImplementation() {
	case proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM:
		result, meta, err = executeCustomFunction(scope, input)
	case proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME:
		if !isMap {
			if input == nil {
				inputsAsMap = make(map[string]any)
			} else {
				return nil, nil, fmt.Errorf("input %v is not in correct format", input)
			}
		}
		result, err = executeRuntimeAction(scope, inputsAsMap)
	case proto.ActionImplementation_ACTION_IMPLEMENTATION_AUTO:
		if !isMap {
			if input == nil {
				inputsAsMap = make(map[string]any)
			} else {
				return nil, nil, fmt.Errorf("input %v is not in correct format", input)
			}
		}
		result, err = executeAutoAction(scope, inputsAsMap)
	default:
		return nil, nil, fmt.Errorf("unhandled unknown action %s of type %s", scope.Action.GetName(), scope.Action.GetImplementation())
	}

	// Generate and send any events for this context.
	// This must run regardless of the action succeeding or failing.
	// Failure to generate events fail silently.
	eventsErr := events.SendEvents(ctx, scope.Schema)
	if eventsErr != nil {
		span.RecordError(eventsErr)
		span.SetStatus(codes.Error, eventsErr.Error())
	}

	return
}

func executeCustomFunction(scope *Scope, inputs any) (any, *common.ResponseMetadata, error) {
	inputsAsMap, _ := inputs.(map[string]any)
	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)
	canAuthoriseEarly, authorised, err := TryResolveAuthorisationEarly(scope, inputsAsMap, permissions)
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

	resp, meta, err := functions.CallFunction(
		scope.Context,
		scope.Action.GetName(),
		inputs,
		permissionState,
	)

	if err != nil {
		return nil, nil, err
	}

	m := &common.ResponseMetadata{
		Headers: http.Header(meta.Headers),
		Status:  meta.Status,
	}

	message := scope.Schema.FindMessage(scope.Action.GetResponseMessageName())

	if asMap, ok := resp.(map[string]any); ok {
		resp, err = transformMessageFileResponses(scope.Context, scope.Schema, message, asMap)
		if err != nil {
			return nil, nil, err
		}
	}

	// For now a custom list function just returns a list of records, but the API's
	// all return an objects containing results and pagination info. So we need
	// to "wrap" the results here.
	// TODO: come up with a better implementation for list functions that can support
	// pagination
	if scope.Action.GetType() == proto.ActionType_ACTION_TYPE_LIST {
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
		}, m, nil
	}

	return resp, m, err
}

func executeRuntimeAction(scope *Scope, inputs map[string]any) (any, error) {
	switch scope.Action.GetName() {
	case requestPasswordResetActionName:
		err := ResetRequestPassword(scope, inputs)
		return map[string]any{}, err
	case passwordResetActionName:
		err := ResetPassword(scope, inputs)
		return map[string]any{}, err
	default:
		return nil, fmt.Errorf("unhandled runtime action: %s", scope.Action.GetName())
	}
}

func executeAutoAction(scope *Scope, inputs map[string]any) (any, error) {
	switch scope.Action.GetType() {
	case proto.ActionType_ACTION_TYPE_GET:
		v, err := Get(scope, inputs)
		// Get() can return nil, but for some reason if we don't explicitly
		// return nil here too the result becomes an empty map, which is rather
		// odd.
		// Simple repo of this: https://play.golang.com/p/MbBzvhrdOm_f
		if v == nil {
			return nil, err
		}
		return v, err
	case proto.ActionType_ACTION_TYPE_UPDATE:
		result, err := Update(scope, inputs)
		return result, err
	case proto.ActionType_ACTION_TYPE_CREATE:
		result, err := Create(scope, inputs)
		return result, err
	case proto.ActionType_ACTION_TYPE_DELETE:
		result, err := Delete(scope, inputs)
		return result, err
	case proto.ActionType_ACTION_TYPE_LIST:
		result, err := List(scope, inputs)
		return result, err
	default:
		return nil, fmt.Errorf("unhandled auto action type: %s", scope.Action.GetType().String())
	}
}
