package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
)

func rewriteNullableInputs(scope *Scope, inputs map[string]any) error {
	var message *proto.Message

	switch scope.Operation.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		message = proto.FindMessage(scope.Schema.Messages, scope.Operation.InputMessageName)
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		rootMessage := proto.FindMessage(scope.Schema.Messages, scope.Operation.InputMessageName)
		valuesField := proto.FindMessageField(rootMessage, "values")
		message = proto.FindMessage(scope.Schema.Messages, valuesField.Type.MessageName.Value)
		if inputs["values"] == nil {
			return nil
		}
		inputs = inputs["values"].(map[string]any)
	case
		proto.OperationType_OPERATION_TYPE_GET,
		proto.OperationType_OPERATION_TYPE_LIST,
		proto.OperationType_OPERATION_TYPE_DELETE,
		proto.OperationType_OPERATION_TYPE_READ,
		proto.OperationType_OPERATION_TYPE_WRITE:
		return nil
	}

	return rewriteNullableInputsInMessage(scope, message, inputs, scope.Model)
}

func rewriteNullableInputsInMessage(scope *Scope, message *proto.Message, inputs map[string]any, currModel *proto.Model) error {

	for key, value := range inputs {
		messageField := proto.FindMessageField(message, key)

		// An 'Any' input field wont be defined in the message
		if messageField == nil {
			continue
		}

		// If the field has a target, then we know it is targeting a field on this model.
		if len(messageField.Target) > 0 {
			modelField := proto.FindField(scope.Schema.Models, messageField.Type.ModelName.Value, key)

			// Determine if this field is optional. If it is, then
			// unwrap it from the nullable type.
			if modelField.Optional {
				var err error
				inputs[key], err = common.ValueFromNullableInput(value)
				if err != nil {
					return common.NewInputValidationError(fmt.Sprintf("invalid value for '%s': %s", key, err.Error()))
				}
			}
			continue
		}

		// If the field is a MESSAGE, then we know we have a nested relationship.
		if messageField.Type.Type == proto.Type_TYPE_MESSAGE {
			// Determine if THIS relationship field is optional. If it is, then
			// unwrap it from the nullable type.
			relationshipField := proto.FindField(scope.Schema.Models, currModel.Name, key)

			if proto.IsHasMany(relationshipField) {
				continue
			}

			if relationshipField.Optional {
				var err error
				inputs[key], err = common.ValueFromNullableInput(value)
				if err != nil {
					return common.NewInputValidationError(fmt.Sprintf("invalid value for '%s': %s", key, err.Error()))
				}
			}

			var asMap map[string]any
			if inputs[key] != nil {
				// TODO: HANDLE IsHasMany relationship
				asMap = inputs[key].(map[string]any)

				// Now rewrite all values within this new message, recursively.
				message := proto.FindMessage(scope.Schema.Messages, messageField.Type.MessageName.Value)
				if relationshipField.Optional {
					valueField := proto.FindMessageField(message, "value")
					message = proto.FindMessage(scope.Schema.Messages, valueField.Type.MessageName.Value)
				}
				nestedModel := proto.FindModel(scope.Schema.Models, messageField.Type.ModelName.Value)
				err := rewriteNullableInputsInMessage(scope, message, asMap, nestedModel)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}
