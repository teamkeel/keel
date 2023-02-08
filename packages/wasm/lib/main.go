//go:build wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/completions"
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
		"validate":    js.FuncOf(validate),
		"format":      js.FuncOf(formatSchema),
		"completions": js.FuncOf(provideCompletions),
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

func provideCompletions(this js.Value, args []js.Value) any {
	return newPromise(func() (any, error) {
		line := args[1].Get("line").Int()
		column := args[1].Get("column").Int()

		completions := completions.Completions(args[0].String(), &node.Position{
			Column: column,
			Line:   line,
		}, args[2].String())

		untypedCompletions := toUntypedArray(completions)

		return map[string]any{
			"completions": js.ValueOf(untypedCompletions),
		}, nil
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

// Type definition for this function:
//
//	validate(schema: string)
func validate(this js.Value, args []js.Value) any {
	return newPromise(func() (any, error) {
		schemaFile := reader.SchemaFile{
			FileName: "schema.keel",
			Contents: args[0].String(),
		}

		builder := schema.Builder{}

		if args[1].Truthy() {
			config, err := config.LoadFromBytes([]byte(args[1].String()))
			if err != nil {
				return nil, err
			}
			builder.Config = config
		}

		_, err := builder.MakeFromInputs(&reader.Inputs{
			SchemaFiles: []reader.SchemaFile{schemaFile},
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
