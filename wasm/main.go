//go:build wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/fatih/color"
	"github.com/teamkeel/keel/runtime/gql"
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

	// A partial ast is returned from parser.Parse if there is a parse error
	// the partial ast will include anything up to the parse error.
	ast, _ := parser.Parse(
		&reader.SchemaFile{
			FileName: "schema.keel",
			Contents: args[0].String(),
		},
	)

	completions := completions.ProvideCompletions(ast, node.Position{
		Column: column,
		Line:   line,
	})

	astMap, _ := toMap(ast)

	var untypedCompletions []any = toUntypedArray(completions)

	return js.ValueOf(
		map[string]any{
			"completions": js.ValueOf(untypedCompletions),
			"ast":         astMap,
		},
	)
}

// getGraphQLSchemas accepts a Keel schema string and returns
// a map of GraphQL schemas where the keys are the names of
// any api's in the Keel schema that used the @graphql attribute
func getGraphQLSchemas(this js.Value, args []js.Value) any {
	return newPromise(func() (any, error) {
		builder := schema.Builder{}

		proto, err := builder.MakeFromInputs(&reader.Inputs{
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

		schemas, err := gql.MakeSchemas(proto)
		if err != nil {
			return nil, err
		}

		res := map[string]any{}
		for _, api := range proto.Apis {
			gqlSchema := schemas[api.Name]
			res[api.Name] = gql.ToSchemaLanguage(*gqlSchema)
		}

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
//   validate(schema: string, options?: {color: boolean})
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
		errs, ok := err.(errorhandling.ValidationErrors)
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
