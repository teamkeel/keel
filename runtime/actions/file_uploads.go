package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

// handleFileUploads will check the inputs for any file uploads for the scope's action and upload them
//
// Currently inline files will be provided as input in a data-url format, we will store these files and change the inputs
// to a structure that will be then saved in the db
func handleFileUploads(scope *Scope, inputs map[string]any) (map[string]any, error) {
	// we handle file uploads for UPDATE and CREATE actions
	if scope.Action.Type != proto.ActionType_ACTION_TYPE_UPDATE && scope.Action.Type != proto.ActionType_ACTION_TYPE_CREATE {
		return inputs, nil
	}
	// check if the values input message for the action has any files
	message := proto.FindValuesInputMessage(scope.Schema, scope.Action.Name)
	if message == nil || !message.HasFiles() {
		return inputs, nil
	}
	storer, err := runtimectx.GetStorage(scope.Context)
	if err != nil {
		return inputs, fmt.Errorf("invalid file storage: %w", err)
	}

	for _, field := range message.Fields {
		// foreach message field that is of inline file type...
		if field.Type != nil && field.Type.Type == proto.Type_TYPE_INLINE_FILE {
			if in, ok := inputs[field.Name]; ok {
				data, ok := in.(string)
				if !ok {
					return inputs, fmt.Errorf("invalid input for field: %s", field.Name)
				}
				// .. we store the file
				file, err := storer.Store(data)
				if err != nil {
					return inputs, fmt.Errorf("storing file: %w", err)
				}

				// ... and then change the input with the file data that should be saved in the db
				fileData, err := file.ToJSON()
				if err != nil {
					return inputs, err
				}
				inputs[field.Name] = fileData
			}
		}
	}

	return inputs, nil
}
