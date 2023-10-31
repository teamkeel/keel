//go:build wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/completions"
	"github.com/teamkeel/keel/schema/definitions"
	"github.com/teamkeel/keel/schema/format"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func init() {
	// we have to declare our functions in an init func otherwise they aren't
	// available in JS land at the call time.
	js.Global().Set("keel", js.ValueOf(map[string]any{
		"validate":      js.FuncOf(validate),
		"format":        js.FuncOf(formatSchema),
		"completions":   js.FuncOf(provideCompletions),
		"getDefinition": js.FuncOf(getDefinition),
	}))
}

func main() {
	done := make(chan bool)
	<-done
}

// newPromise wraps the provided function in a Javascript Promise
// and returns that promise. It then either resolves or rejects
// the promise based on whether fn returns an error or not
func newPromise(fn func() (any, error)) any {
	handler := js.FuncOf(func(this js.Value, args []js.Value) any {
		resolve := args[0]
		reject := args[1]

		go func() {
			// handle panics
			defer func() {
				if r := recover(); r != nil {
					msg := "panic"
					switch r.(type) {
					case string:
						msg = r.(string)
					case error:
						e := r.(error)
						msg = e.Error()
					}
					// err should be an instance of `error`, eg `errors.New("some error")`
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(msg)
					reject.Invoke(errorObject)
				}
			}()

			data, err := fn()
			if err != nil {
				// err should be an instance of `error`, eg `errors.New("some error")`
				errorConstructor := js.Global().Get("Error")
				errorObject := errorConstructor.New(err.Error())
				reject.Invoke(errorObject)
			} else {
				resolve.Invoke(js.ValueOf(data))
			}
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

// Expected argument to definitions API:
//
//	{
//		position: {
//			filename: "",
//			line: 1,
//			column: 1,
//		},
//		schemaFiles: [
//			{
//				filename: "",
//				contents: "",
//			},
//		],
//		config: "",
//	}
func provideCompletions(this js.Value, args []js.Value) any {
	return newPromise(func() (any, error) {
		positionArg := args[0].Get("position")
		pos := &node.Position{
			Filename: positionArg.Get("filename").String(),
			Line:     positionArg.Get("line").Int(),
			Column:   positionArg.Get("column").Int(),
		}

		schemaFilesArg := args[0].Get("schemaFiles")
		schemaFiles := []*reader.SchemaFile{}
		for i := 0; i < schemaFilesArg.Length(); i++ {
			f := schemaFilesArg.Index(i)
			schemaFiles = append(schemaFiles, &reader.SchemaFile{
				FileName: f.Get("filename").String(),
				Contents: f.Get("contents").String(),
			})
		}

		configSrc := args[0].Get("config")
		var cfg *config.ProjectConfig
		if configSrc.Truthy() {
			// We don't care about errors here, if we can get a config object
			// back we'll use it, if not then we'll run validation without it
			cfg, _ = config.LoadFromBytes([]byte(configSrc.String()))
		}

		completions := completions.Completions(schemaFiles, pos, cfg)

		untypedCompletions := toUntypedArray(completions)

		return map[string]any{
			"completions": js.ValueOf(untypedCompletions),
		}, nil
	})
}

// Expected argument to definitions API:
//
//	{
//		position: {
//			filename: "",
//			line: 1,
//			column: 1,
//		},
//		schemaFiles: [
//			{
//				filename: "",
//				contents: "",
//			},
//		],
//	}
func getDefinition(this js.Value, args []js.Value) any {
	return newPromise(func() (any, error) {
		positionArg := args[0].Get("position")
		pos := definitions.Position{
			Filename: positionArg.Get("filename").String(),
			Line:     positionArg.Get("line").Int(),
			Column:   positionArg.Get("column").Int(),
		}

		schemaFilesArg := args[0].Get("schemaFiles")
		schemaFiles := []*reader.SchemaFile{}
		for i := 0; i < schemaFilesArg.Length(); i++ {
			f := schemaFilesArg.Index(i)
			schemaFiles = append(schemaFiles, &reader.SchemaFile{
				FileName: f.Get("filename").String(),
				Contents: f.Get("contents").String(),
			})
		}

		def := definitions.GetDefinition(schemaFiles, pos)
		if def == nil {
			return nil, nil
		}

		return toMap(def)
	})
}

func formatSchema(this js.Value, args []js.Value) any {
	return newPromise(func() (any, error) {
		src := args[0].String()
		ast, err := parser.Parse(&reader.SchemaFile{
			FileName: "schema.keel",
			Contents: src,
		})
		if err != nil {
			// if the schema can't be parsed then just return it as-is
			return src, nil
		}

		return format.Format(ast), nil
	})
}

// Expected argument to validate API:
//
//	{
//		schemaFiles: [
//			{
//				filename: "",
//				contents: "",
//			},
//		],
//		config: "<YAML config file>"
//	}
//
// The config file source is optional.
func validate(this js.Value, args []js.Value) any {
	return newPromise(func() (any, error) {

		schemaFilesArg := args[0].Get("schemaFiles")
		schemaFiles := []*reader.SchemaFile{}
		for i := 0; i < schemaFilesArg.Length(); i++ {
			f := schemaFilesArg.Index(i)
			schemaFiles = append(schemaFiles, &reader.SchemaFile{
				FileName: f.Get("filename").String(),
				Contents: f.Get("contents").String(),
			})
		}

		builder := schema.Builder{}

		configSrc := args[0].Get("config")
		if configSrc.Truthy() {
			// We don't care about errors here, if we can get a config object
			// back we'll use it, if not then we'll run validation without it
			config, _ := config.LoadFromBytes([]byte(configSrc.String()))
			if config != nil {
				builder.Config = config
			}
		}

		_, err := builder.MakeFromInputs(&reader.Inputs{
			SchemaFiles: schemaFiles,
		})

		if err != nil {
			errs, ok := err.(*errorhandling.ValidationErrors)
			if !ok {
				return nil, err
			}

			validationErrors, err := toMap(errs)
			if err != nil {
				return nil, err
			}

			return validationErrors, nil
		}

		return map[string]any{
			"errors": []any{},
		}, nil
	})
}

// js.ValueOf can only marshall map[string]any to a JS object
// so for structs we need to do the struct->json->map[string]any dance
func toMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var res map[string]any
	err = json.Unmarshal(b, &res)
	return res, err
}

func toUntypedArray(items []*completions.CompletionItem) (i []any) {
	for _, item := range items {
		b, err := json.Marshal(item)

		if err != nil {
			continue
		}
		var res any
		err = json.Unmarshal(b, &res)

		if err != nil {
			continue
		}
		i = append(i, res)
	}

	return i
}
