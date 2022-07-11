package gql

import (
	"context"
	"encoding/json"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"gorm.io/gorm"
)

// A Handler implements a request handler for client GraphQL requests, using
// a GraphQL schema that it holds internally. The constructor(s) for Handler,
// allow it to generate its GraphQL schema automatically - with reference to
// a given proto.Api.
//
// It is designed to be good for incorporating into any type
// of server - because it is decoupled from HTTP, and from requests and responses being
// wrapped in JSON objects.
//
// You construct a set of Handler(s) by calling NewHandlers, or NewHandlersFromJSON.
//
// You ask a Handler to respond to a GraphQL request by calling its Handle method.
type Handler struct {
	gSchema *graphql.Schema
	gormDB  *gorm.DB
}

// Handle executes the given GraphQL request against the GraphQL schema that is
// held internally by this handler. You pass in the query as a plain
// GraphQL query string (i.e. not JSON).
func (h *Handler) Handle(gqlQuery string) (result *graphql.Result) {

	result = graphql.Do(graphql.Params{
		Schema:         *h.gSchema,
		Context:        context.Background(),
		RequestString:  gqlQuery,
		VariableValues: map[string]any{},
	})
	return result
}

// NewHandlers provide a Handler for each of the APIs specified in the
// given proto.Schema. These are returned in a map, keyed on the API names.
// The mapping is intended to make it easy for a client to register each of the
// handlers at individual endpoints derived from the API name. E.g. "/graphql/api-1"
func NewHandlers(pSchema *proto.Schema, gormDB *gorm.DB) (map[string]*Handler, error) {
	gSchemaMaker := newMaker(pSchema, gormDB)
	gSchemas, err := gSchemaMaker.make()
	if err != nil {
		return nil, err
	}
	handlers := map[string]*Handler{}
	for apiName, s := range gSchemas {
		handler := &Handler{
			gSchema: s,
			gormDB:  gormDB,
		}
		handlers[apiName] = handler
	}
	return handlers, nil
}

// NewHandlersFromJSON is a variation on NewHandler. It provides a Handler for each of the APIs
// specified in the
// given (JSON serialized) proto.Schema.
// These are returned in a map, keyed on the API names.
// The mapping is intended to make it easy for a client to register each of the
// handlers at individual endpoints derived from the API name. E.g. "/graphql/api-1"
func NewHandlersFromJSON(pSchemaJSON string, gormDB *gorm.DB) (map[string]*Handler, error) {
	pSchema := proto.Schema{}
	err := json.Unmarshal([]byte(pSchemaJSON), &pSchema)
	if err != nil {
		return nil, err
	}
	return NewHandlers(&pSchema, gormDB)
}
