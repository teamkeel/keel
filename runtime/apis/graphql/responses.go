package graphql

import (
	"fmt"
)

// connectionResponse consumes the raw records returned by actions.List() (and similar),
// and wraps them into a Node+Edges structure that is good for the connections pattern
// return type and is expected by the GraphQL schema for the List action.
// See https://relay.dev/graphql/connections.htm
func connectionResponse(data map[string]any) (resp map[string]any, err error) {
	results := []map[string]any{}

	// From custom functions.
	r, ok := data["results"].([]any)
	if ok {
		for _, v := range r {
			value, _ := v.(map[string]any)
			results = append(results, value)
		}
	} else {
		// From built-in ops.
		results, ok = data["results"].([]map[string]any)
		if !ok {
			return nil, fmt.Errorf("list result does not contain results keys")
		}
	}

	pageInfo := data["pageInfo"].(map[string]any)

	edges := []map[string]any{}
	for _, record := range results {
		edge := map[string]any{
			"cursor": record["id"],
			"node":   record,
		}
		edges = append(edges, edge)
	}

	resp = map[string]any{
		"pageInfo": pageInfo,
		"edges":    edges,
	}

	if data["resultInfo"] != nil {
		resp["resultInfo"] = data["resultInfo"]
	}

	return resp, nil
}
