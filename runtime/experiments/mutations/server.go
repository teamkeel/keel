/*
This is iteration zero of a GraphQL server that will consume a proto.Schema at
run time, and implement the GraphQL contract dynamically by building an
"executable schema" on the fly programmatically using the GraphQL SDK.

This iteration is based on the official graphql-go/graphql introductory example.
It does only a Mutation operation, the schema construction is hard coded,

As it stands, it serves two purposes:
1) To illustrate the 101 of how you use the SDK to compose an executable schema.
2) For me to probe and better understand the minimum necessary coupling between
   the components that is required for it to work. See comments within to explain what
   I mean by that.
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"
)

func main() {
	s := NewServer()
	s.ListenAndServe()
}

func NewServer() *http.Server {

	// Compose the types that form the GQL hierarchy - in dependency order, i.e. bottom up.
	// I've stripped the user type down to the smallest possible - it just has a name field.
	var todoType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Todo",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"text": &graphql.Field{
				Type: graphql.String,
			},
			"done": &graphql.Field{
				Type: graphql.Boolean,
			},
		},
	})

	// The root level mutation
	var rootMutation = graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			/*
				curl -g 'http://localhost:8080/graphql?query=mutation+_{createTodo(text:"My+new+todo"){id,text,done}}'
			*/
			"createTodo": &graphql.Field{
				Type:        todoType, // the return type for this field
				Description: "Create new todo",
				// createTodo takes a message (string) argument.
				Args: graphql.FieldConfigArgument{
					"text": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					text, _ := params.Args["text"].(string)

					newID := "999" // doesn't matter for test purposes

					// Construct the plain-go object to represent the new object.
					newTodo := Todo{
						ID:   newID,
						Text: text,
						Done: false,
					}

					// This is where the real database create operation would be done - using
					// the plain-go type object.

					// We can return the plain-go object because it maps structurally and with
					// tag names to the required todoType.
					return newTodo, nil
				},
			},
		},
	})

	// For some reason you must also compose the schema with a query as well as a mutation.
	var rootQuery = graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"todo": &graphql.Field{
				Type:        todoType,
				Description: "Get single todo",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					// return a fake one - we're not interested in queries here.
					return Todo{}, nil
				},
			},
		},
	})

	// Now we can compose a GQL schema that uses the root level queryType we just built.
	var schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    rootQuery,
			Mutation: rootMutation,
		},
	)

	// Register a handler with the default server MUX to handle requests to /graphql.
	// It can be handed off directly (with JSON wrapping of input and output) to graphql.Do.
	http.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
		var p postData
		if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
			w.WriteHeader(400)
			return
		}
		result := graphql.Do(graphql.Params{
			Context:        context.Background(),
			Schema:         schema,
			RequestString:  p.Query,
			VariableValues: p.Variables,
			OperationName:  p.Operation,
		})
		if err := json.NewEncoder(w).Encode(result); err != nil {
			fmt.Printf("could not write result to response: %s", err)
		}
	})

	fmt.Println("Now server is running on port 8080")

	s := &http.Server{
		Addr: ":8080",
	}
	return s
}

type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

type postData struct {
	Query     string                 `json:"query"`
	Operation string                 `json:"operation"`
	Variables map[string]interface{} `json:"variables"`
}
