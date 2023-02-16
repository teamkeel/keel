package graphql

import "fmt"

// connectionResponse consumes the raw records returned by actions.List() (and similar),
// and wraps them into a Node+Edges structure that is good for the connections pattern
// return type and is expected by the GraphQL schema for the List operation.
// See https://relay.dev/graphql/connections.htm
func connectionResponse(data map[string]any) (resp any, err error) {
	results, ok := data["results"].([]map[string]any)
	if !ok {
		return nil, fmt.Errorf("list result does not contain results keys")
	}

	hasNextPage, _ := data["hasNextPage"].(bool)

	startCursor := data["startCursor"].(string)
	endCursor := data["endCursor"].(string)

	edges := []map[string]any{}

	for _, record := range results {
		edge := map[string]any{
			"cursor": record["id"],
			"node":   record,
		}
		edges = append(edges, edge)
	}

	pageInfo := map[string]any{
		"hasNextPage": hasNextPage,
		"startCursor": startCursor,
		"endCursor":   endCursor,
	}
	resp = map[string]any{
		"pageInfo": pageInfo,
		"edges":    edges,
	}
	return resp, nil
}
