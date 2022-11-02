//go:build wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"syscall/js"

	"github.com/fatih/color"
	"github.com/graphql-go/graphql/testutil"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/apis/graphql"
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
		"validate":          js.FuncOf(validate),
		"format":            js.FuncOf(formatSchema),
		"completions":       js.FuncOf(provideCompletions),
		"getGraphQLSchemas": js.FuncOf(getGraphQLSchemas),
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
	line := args[1].Get("line").Int()
	column := args[1].Get("column").Int()

	completions := completions.Completions(args[0].String(), &node.Position{
		Column: column,
		Line:   line,
	})

	untypedCompletions := toUntypedArray(completions)

	return js.ValueOf(
		map[string]any{
			"completions": js.ValueOf(untypedCompletions),
		},
	)
}

// getGraphQLSchemas accepts a Keel schema string and returns
// a map of GraphQL schemas where the keys are the names of
// any api's in the Keel schema that used the @graphql attribute
func getGraphQLSchemas(this js.Value, args []js.Value) any {

	return newPromise(func() (any, error) {
		builder := schema.Builder{}

		protoSchema, err := builder.MakeFromInputs(&reader.Inputs{
			SchemaFiles: []reader.SchemaFile{
				{
					FileName: "schema.keel",
					Contents: args[0].String(),
				},
			},
		})
		if err != nil {
			return nil, err
		}

		res := map[string]any{}

		var api *proto.Api
		for _, v := range protoSchema.Apis {
			if v.Type == proto.ApiType_API_TYPE_GRAPHQL {
				api = v
				break
			}
		}

		if api == nil {
			return res, nil
		}

		handler := runtime.NewHandler(protoSchema)

		body, err := json.Marshal(map[string]string{
			"query": testutil.IntrospectionQuery,
		})
		if err != nil {
			return nil, err
		}

		response, err := handler(&http.Request{
			URL: &url.URL{
				Path: "/" + api.Name,
			},
			Method: http.MethodPost,
			Body:   io.NopCloser(bytes.NewReader(body)),
		})

		if response.Status != 200 {
			return nil, fmt.Errorf("error introspecting graphql schema: %s", response.Body)
		}

		res[api.Name] = graphql.ToGraphQLSchemaLanguage(response)
		return res, nil
	})
}

func formatSchema(this js.Value, args []js.Value) any {
	ast, err := parser.Parse(&reader.SchemaFile{
		FileName: "schema.keel",
		Contents: args[0].String(),
	})
	if err != nil {
		// if the schema can't be parsed then just return it as-is
		return js.ValueOf(args[0].String())
	}

	return js.ValueOf(format.Format(ast))
}

// Type definition for this function:
//
//	validate(schema: string, options?: {color: boolean})
func validate(this js.Value, args []js.Value) any {

	if len(args) > 1 {
		withColor := args[1].Get("color").Truthy()
		if withColor {
			color.NoColor = false
		}
	}

	schemaFile := reader.SchemaFile{
		FileName: "schema.keel",
		Contents: args[0].String(),
	}

	builder := schema.Builder{}
	var validationErrors map[string]any
	var validationOutput string

	_, err := builder.MakeFromInputs(&reader.Inputs{
		SchemaFiles: []reader.SchemaFile{schemaFile},
	})

	if err != nil {
		errs, ok := err.(*errorhandling.ValidationErrors)
		if ok {
			validationErrors, err = toMap(errs)
			if err != nil {
				return js.ValueOf(map[string]any{
					"error": err.Error(),
				})
			}

			validationOutput, err = errs.ToAnnotatedSchema([]reader.SchemaFile{
				schemaFile,
			})
			if err != nil {
				return js.ValueOf(map[string]any{
					"error": err.Error(),
				})
			}
		} else {
			return js.ValueOf(map[string]any{
				"error": err.Error(),
			})
		}
	}

	var astMap map[string]any
	asts := builder.ASTs()
	if len(asts) > 0 {
		astMap, err = toMap(asts[0])
		if err != nil {
			return js.ValueOf(map[string]any{
				"error": err.Error(),
			})
		}
	}

	return js.ValueOf(map[string]any{
		"ast":              astMap,
		"validationErrors": validationErrors,
		"validationOutput": validationOutput,
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
