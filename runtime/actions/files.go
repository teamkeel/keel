package actions

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/storage"
)

// handleFileUploads will check the inputs for any file uploads for the scope's action and upload them
//
// Currently inline files will be provided as input in a data-url format, we will store these files and change the inputs
// to a structure that will be then saved in the db.
func handleFileUploads(scope *Scope, inputs map[string]any) (map[string]any, error) {
	// we handle file uploads for UPDATE and CREATE actions
	if scope.Action.GetType() != proto.ActionType_ACTION_TYPE_UPDATE && scope.Action.GetType() != proto.ActionType_ACTION_TYPE_CREATE {
		return inputs, nil
	}
	// check if the values input message for the action has any files
	message := proto.FindValuesInputMessage(scope.Schema, scope.Action.GetName())
	if message == nil {
		return inputs, nil
	}

	storer, err := runtimectx.GetStorage(scope.Context)
	if err != nil {
		return inputs, fmt.Errorf("invalid file storage: %w", err)
	}

	for _, field := range message.GetFields() {
		// foreach message field that is of inline file type...
		if field.GetType() != nil && field.GetType().GetType() == proto.Type_TYPE_FILE {
			if in, ok := inputs[field.GetName()]; ok {
				// null files don't need uploading
				if in == nil {
					continue
				}

				if field.GetType().GetRepeated() {
					data, ok := in.([]string)
					if !ok {
						return inputs, fmt.Errorf("invalid input for field: %s", field.GetName())
					}

					fileInfos := []any{}
					for _, d := range data {
						fi, err := storer.Store(d)
						if err != nil {
							return inputs, fmt.Errorf("storing file: %w", err)
						}

						fileInfos = append(fileInfos, fi)
					}
					inputs[field.GetName()] = fileInfos
				} else {
					data, ok := in.(string)
					if !ok {
						return inputs, fmt.Errorf("invalid input for field: %s", field.GetName())
					}

					fi, err := storer.Store(data)
					if err != nil {
						return inputs, fmt.Errorf("storing file: %w", err)
					}

					inputs[field.GetName()] = fi
				}
			}
		}
	}

	return inputs, nil
}

// transformModelFileResponses will take the results for the given scope's action execution and parse and transform the file responses.
func transformModelFileResponses(ctx context.Context, model *proto.Model, results map[string]any) (map[string]any, error) {
	if model == nil {
		return results, nil
	}

	store, err := runtimectx.GetStorage(ctx)
	if err != nil {
		return results, fmt.Errorf("no file storage implementation: %w", err)
	}

	for _, field := range model.FileFields() {
		if data, found := results[field.GetName()]; found && data != nil {
			if field.GetType().GetRepeated() {
				fis, ok := data.([]storage.FileInfo)
				if !ok {
					return results, fmt.Errorf("invalid response for field: %s", field.GetName())
				}

				resp := []storage.FileResponse{}
				for _, fi := range fis {
					// now we're hydrating the db file info with data from our storage service
					// e.g. injecting signed URLs for direct file downloads
					hydrated, err := store.GenerateFileResponse(&fi)
					if err != nil {
						return results, fmt.Errorf("failed hydrating file data: %w", err)
					}

					resp = append(resp, hydrated)
				}
				results[field.GetName()] = resp
			} else {
				fi, ok := data.(storage.FileInfo)
				if !ok {
					return results, fmt.Errorf("invalid response for field: %s", field.GetName())
				}

				// now we're hydrating the db file info with data from our storage service
				// e.g. injecting signed URLs for direct file downloads
				resp, err := store.GenerateFileResponse(&fi)
				if err != nil {
					return results, fmt.Errorf("failed hydrating file data: %w", err)
				}
				results[field.GetName()] = resp
			}
		}
	}

	return results, nil
}

// transformMessageFileResponses will take the results from the functions runtime and parse and transform the file responses.
func transformMessageFileResponses(ctx context.Context, schema *proto.Schema, message *proto.Message, results map[string]any) (map[string]any, error) {
	if message == nil {
		return results, nil
	}

	for _, field := range message.GetFields() {
		if v, found := results[field.GetName()]; found && v != nil {
			switch field.GetType().GetType() {
			case proto.Type_TYPE_MESSAGE:

				nested := schema.FindMessage(field.GetType().GetMessageName().GetValue())

				var err error
				if field.GetType().GetRepeated() {
					arr := v.([]any)
					for i, el := range arr {
						arr[i], err = transformMessageFileResponses(ctx, schema, nested, el.(map[string]any))
						if err != nil {
							return nil, err
						}
					}
					results[field.GetName()] = arr
				} else {
					results[field.GetName()], err = transformMessageFileResponses(ctx, schema, nested, v.(map[string]any))
					if err != nil {
						return nil, err
					}
				}

			case proto.Type_TYPE_FILE:
				data, ok := v.(map[string]any)
				if !ok {
					return results, fmt.Errorf("invalid response for field: %s", field.GetName())
				}

				fi := storage.FileInfo{
					Key:         data["key"].(string),
					Filename:    data["filename"].(string),
					ContentType: data["contentType"].(string),
					Size:        int(data["size"].(float64)),
				}

				// now we're hydrating the db file info with data from our storage service if we have one
				// e.g. injecting signed URLs for direct file downloads
				if store, err := runtimectx.GetStorage(ctx); err == nil {
					hydrated, err := store.GenerateFileResponse(&fi)
					if err != nil {
						return results, fmt.Errorf("failed retrieve hydrated file data: %w", err)
					}
					results[field.GetName()] = hydrated
				} else {
					// or, we don't have a storage service so we can just return the data saved in the db
					results[field.GetName()] = fi
				}
			}
		}
	}
	return results, nil
}
