/*
This is iteration zero of a GraphQL server that will consume a proto.Schema at
run time, and implement the GraphQL contract dynamically by building an
"executable schema" on the fly programmatically using the GraphQL SDK.

This iteration is based on the official graphql-go/graphql introductory example.
It does only a Query operation, the schema construction is hard coded, and the
resolvers work by accessing a trivial in-memory data structure.

As it stands, it serves two purposes:
1) To illustrate the 101 of how you use the SDK to compose an executable schema.
2) For me to probe and better understand the minimum necessary coupling between
   the components that is required for it to work. See comments within to explain what
   I mean by that.
*/

package main

import (
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
	var userType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "User",
			Fields: graphql.Fields{
				"name": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)

	// Now the root level type - which is a query, that has-a userType field.
	// Note that it defines an Arg for the query called <id> of type string.
	// It seems the Arg need not refer to a field in the object type (I took it out),
	// but is only part of the resolver contract.
	var queryType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"user": &graphql.Field{
					Type: userType,
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					// The resolver for the <user> field harvests the <id> argument from an incoming query,
					// and uses it to look up the data from some scratch built in data.
					// So the scope of the Arg is resolver land - and needn't be connected with the
					// field's type.
					//
					// Seems a resolver for a non-scalar field can return any old thing. The GQL implementation will
					// then recurse for the expected fields within. It seems that if it expects a field within called
					// "foo", and the object doesn't have a field called (or JSON tagged) as "foo" - you just get
					// nil for that field's value. (Presumably unless you define it as non-nullable).
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						idQuery, isOK := p.Args["id"].(string)
						if isOK {
							// The resolver is going to return whatever (arbitrary) type this lookup evaluates to
							// to - note the resolver function signature returns a nil-interface.
							// In our case it will be an instance of type arbitaryType.
							return psuedoDatabase[idQuery], nil
						}
						return nil, nil
					},
				},
			},
		})

	// Now we can compose a GQL schema that uses the root level queryType we just built.
	var schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query: queryType,
		},
	)

	_ = schema

	// Register a handler with the default server MUX to handle requests to /graphql by first calling
	// the executeQuery function - passing in the executable schema. Then json encoding the result
	// to form the response.
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		// Delegate to the executeQuery function we defined above.
		// Note the use of standard HTTP package URL.Query().Get(<paramName>), where
		// the param name in this case is "query" to reflect the url encoded query in the curl
		// request example below.
		result := executeQuery(r.URL.Query().Get("query"), schema)
		// Standard JSON encoding of the response - (ignoring the potential error return value)
		json.NewEncoder(w).Encode(result)
	})

	// And finally start the server.
	fmt.Println("Now server is running on port 8080")
	fmt.Println("Test with Get      : curl -g 'http://localhost:8080/graphql?query={user(id:\"1\"){name}}'")
	fmt.Println(`You should get this back: {"data":{"user":{"name":"fred"}}}`)

	s := &http.Server{
		Addr: ":8080",
	}
	return s
}

// A pure-go handler function that receives a query (string), and an
// executable GQL Schema. A trivial wrapper over graphql.Do
func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}
	return result
}

// Note I've used arbitrary names for types and fields here that are different
// from those used by the GQL Schema.
type arbitaryType struct {
	// Note I've not given it an <id> field. By definition it cannot be needed by
	// GQL despite being defined as an Arg for the user field, because GQL
	// has no knowledge of this structure type.

	// ID: string

	// The implementation of the resolver for the "user" field in the query, does a lookup
	// in our psuedoDatabase, and will find and return an object of arbitaryType. That will
	// completely satisfy the resolve repsonsibility it seems. I.e. the returned object
	// will not be scrutinized for comformance with the GQL userType specified at that stage.
	// It seems it will then later recurse to resolvers for each field therein (by definition until
	// it reaches a scalar field) - and then resolve the scalar field automatically by finding a field
	// in the arbitraryType object that has a JSON tage of the right name (and has the right type).
	ArbitaryField string `json:"name"`
}

var psuedoDatabase map[string]arbitaryType = map[string]arbitaryType{
	"1": {
		ArbitaryField: "fred",
	},
}

// Summary of the minimum binding requirements discovered.
//
// If you define a GQL type that has a field with the name "xxx", and include that
// field in a query http request, then the resolver for that field, will look in the
// arbitrary object it is given, for a field whose JSON tag is also "xxx".
// If the object doesn't have such a field - the resolver shown above survives but
// populates the field in the response with a nil.
//
// This happens at the level of the graphql.Do() level - nothing to do the
// JSON encoding of the http response.
//
// So it seems the JSON tag is part of the graphql.Do() contract?
// It maybe it would also be satisfied with the field actually having the right
// name - but that means the filed would then have to be exported, which clashes
// with the lower case GQL field name convention.
