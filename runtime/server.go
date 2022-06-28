package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/gql"
)

// NewServer returns an http.Server that implements the GraphQL API(s)
// implied by the given keel schema.
//
// Each API from the Keel Schema is served at /graphql/<api-name>.
//
// It is left to the client to call ListenAndServe() on it.
// And to call Shutdown() on it when done with it.
//
// You effectively pass in a proto.Schema, but it accepts a
// JSON serialised form for that to make it suitable be be delivered
// from a deployment script.
//
// TODO have the entry point function pick up the schema from an environment variable,
// then delegate to an internal function that accepts a string as an argument.
func NewServer(schemaProtoJSON string) (*http.Server, error) {
	var schema proto.Schema
	if err := json.Unmarshal([]byte(schemaProtoJSON), &schema); err != nil {
		return nil, fmt.Errorf("error unmarshalling the schema: %v", err)
	}
	gSchemaMaker := gql.NewMaker(&schema)
	gSchemas, err := gSchemaMaker.Make()
	if err != nil {
		return nil, err
	}

	for apiName, gSchema := range gSchemas {
		handler, err := newHandler(gSchema)
		if err != nil {
			panic(err.Error())
		}
		serveAt := fmt.Sprintf("/graphql/%s", apiName)
		http.Handle(serveAt, handler)
	}
	s := &http.Server{
		Addr: ":8080",
	}
	return s, nil
}

// A handler is an HTTP request handler, that expects incoming requests to
// contain a GraphQL query embedded in a JSON object. It replies with the
// results of executing that query against the graphql.Schema passed to the
// handler at construction time.
type handler struct {
	schema *graphql.Schema
}

func newHandler(gSchema *graphql.Schema) (*handler, error) {
	return &handler{schema: gSchema}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		// todo - how to surface the error details?
		fmt.Printf("XXXX request is malformed json: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("XXXX json decoded params is: %+v\n", params)
	fmt.Printf("XXXX isolated query string is: %s\n", params.Query)

	result := graphql.Do(graphql.Params{
		Schema:         *h.schema,
		Context:        context.Background(),
		RequestString:  params.Query,
		VariableValues: params.Variables,
	})

	responseJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("XXXX json response:\n%s\n", string(responseJSON))
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}
