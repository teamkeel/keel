package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/gql"
	"gorm.io/gorm"
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
// For the GraphQL heaving lifting it uses the handlers provided by /runtime/gql/.
func NewServer(schema *proto.Schema, gormDB *gorm.DB) (*http.Server, error) {
	plainHandlers, err := gql.NewHandlers(schema, gormDB)
	if err != nil {
		return nil, err
	}

	for apiName, gqlHandler := range plainHandlers {
		httpHandler, err := newHTTPHandler(gqlHandler)
		if err != nil {
			panic(err.Error())
		}
		serveAt := fmt.Sprintf("/graphql/%s", apiName)
		http.Handle(serveAt, httpHandler)
	}
	s := &http.Server{
		Addr: ":8080",
	}
	return s, nil
}

// A handler is an HTTP request handler, that expects incoming requests to
// contain a GraphQL query embedded in a JSON object. It is a simple wrapper
// over a gql.Handler that adds only parsing the incoming JSON request,
// and returning the result wrapped in JSON.
type handler struct {
	// gqlHandler is the (non HTTP) GraphQL handler that it delegates to.
	gqlHandler *gql.Handler
}

func newHTTPHandler(gqlHandler *gql.Handler) (*handler, error) {
	return &handler{gqlHandler: gqlHandler}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Unpack the JSON request to access the GraphQL query string.
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		// todo - how to surface the error details?
		fmt.Printf("error request is malformed json: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Delegate all the heavy lifting to a plain (non HTTP) handler.
	result := h.gqlHandler.Handle(params.Query)

	// And JSON encode the response.
	responseJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}
