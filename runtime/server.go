package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/makegql"
)

// NewServer returns an http.Server that implements the GraphQL API as
// implied by the given keel schema.
//
// It is left to the client to call ListenAndServe() on it.
// And to call Shutdown() on it when done with it.
//
// You effectively pass in a proto.Schema, but it accepts a
// JSON serialised form for that to make it suitable be be delivered
// from a deployment script.
func NewServer(schemaProtoJSON string) (*http.Server, error) {
	var schema proto.Schema
	if err := json.Unmarshal([]byte(schemaProtoJSON), &schema); err != nil {
		return nil, fmt.Errorf("error unmarshalling the schema: %v", err)
	}
	gqlSchemas := makegql.MakeGQLSchemas(&schema)
	_ = gqlSchemas

	return nil, nil
}
