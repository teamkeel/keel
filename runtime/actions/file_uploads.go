package actions

import (
	"encoding/json"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/storage"
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
				// .. we store the fi
				fi, err := storer.Store(data)
				if err != nil {
					return inputs, fmt.Errorf("storing file: %w", err)
				}

				// ... and then change the input with the file data that should be saved in the db
				fileInfo, err := fi.ToJSON()
				if err != nil {
					return inputs, err
				}
				inputs[field.Name] = fileInfo
			}
		}
	}

	return inputs, nil
}

// transformFileResponses will take the results for the given scope's action execution and parse and transform the file responses
func transformFileResponses(scope *Scope, results map[string]any) (map[string]any, error) {
	model := proto.FindModel(scope.Schema.Models, scope.Action.ModelName)
	if model == nil {
		return results, nil
	}

	for _, field := range model.FileFields() {
		if fileJSON, found := results[field.Name]; found {
			data, ok := fileJSON.(string)
			if !ok {
				return results, fmt.Errorf("invalid response for field: %s", field.Name)
			}

			fi := storage.FileInfo{}
			if err := json.Unmarshal([]byte(data), &fi); err != nil {
				return results, fmt.Errorf("failed to unmarshal file data: %w", err)
			}

			// now we're hydrating the db file info with data from our storage service if we have one
			// e.g. injecting signed URLs for direct file downloads
			if store, err := runtimectx.GetStorage(scope.Context); err == nil {
				hydrated, err := store.HydrateFileInfo(&fi)
				if err != nil {
					return results, fmt.Errorf("failed retrieve hydrated file data: %w", err)
				}
				results[field.Name] = hydrated
			} else {
				// or, we don't have a storage service so we can just return the data saved in the db
				results[field.Name] = fi
			}
		}
	}
	return results, nil
}
