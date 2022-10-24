package graphql

import (
	"fmt"
	"time"

	"github.com/bykof/gostradamus"
	"github.com/graphql-go/graphql"
	"github.com/nleeper/goment"
	"github.com/teamkeel/keel/proto"
)

var protoTypeToGraphQLQueryInput = map[proto.Type]graphql.Input{
	proto.Type_TYPE_ID:        idQueryInputType,
	proto.Type_TYPE_STRING:    stringQueryInputType,
	proto.Type_TYPE_INT:       intQueryInputType,
	proto.Type_TYPE_BOOL:      booleanQueryInput,
	proto.Type_TYPE_TIMESTAMP: timestampQueryInputType,
	proto.Type_TYPE_DATETIME:  timestampQueryInputType,
	proto.Type_TYPE_DATE:      dateQueryInputType,
}

var deleteResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteResponse",
	Fields: graphql.Fields{
		"success": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Boolean),
		},
	},
})

var pageInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PageInfo",
	Fields: graphql.Fields{
		"hasNextPage": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Boolean),
		},
		"startCursor": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"endCursor": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"totalCount": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
		},
	},
})

var fromNowType = graphql.Field{
	Name: "fromNow",
	Type: graphql.NewNonNull(graphql.String),
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		t, ok := p.Source.(time.Time)

		if !ok {
			return nil, fmt.Errorf("not a valid time")
		}

		g, err := goment.New(t)

		if err != nil {
			return nil, err
		}

		return g.FromNow(), nil
	},
}

var formattedDateType = &graphql.Field{
	Name:        "formatted",
	Description: "Formatted timestamp. Uses standard datetime formats",
	Type:        graphql.NewNonNull(graphql.String),
	Args: graphql.FieldConfigArgument{
		"format": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		t, ok := p.Source.(time.Time)

		if !ok {
			return nil, fmt.Errorf("not a valid time")
		}

		formatArg, ok := p.Args["format"].(string)

		if !ok {
			return nil, fmt.Errorf("no format argument provided")
		}

		// Go prefers to use layout as the basis for date formats
		// However most users of the api will likely be used to date
		// formats such as YYYY-mm-dd so therefore the library below
		// provides a conversion inbetween standard date formats and go's
		// layout format system
		// Format spec: https://github.com/bykof/gostradamus/blob/master/formatting.go#L11-L42
		dateTime := gostradamus.DateTimeFromTime(t)

		return dateTime.Format(formatArg), nil
	},
}

var timestampType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Timestamp",
	Fields: graphql.Fields{
		"seconds": &graphql.Field{
			Name:        "seconds",
			Description: "Seconds since unix epoch",
			Type:        graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				t, ok := p.Source.(time.Time)

				if !ok {
					return nil, fmt.Errorf("not a valid time")
				}

				return t.Unix(), nil
			},
		},
		"year": &graphql.Field{
			Name: "year",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				d, ok := p.Source.(time.Time)

				if !ok {
					return nil, fmt.Errorf("not a valid date")
				}

				return d.Year(), nil
			},
		},
		"month": &graphql.Field{
			Name: "month",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				d, ok := p.Source.(time.Time)

				if !ok {
					return nil, fmt.Errorf("not a valid date")
				}

				return int(d.Month()), nil
			},
		},
		"day": &graphql.Field{
			Name: "day",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				d, ok := p.Source.(time.Time)

				if !ok {
					return nil, fmt.Errorf("not a valid date")
				}

				return d.Day(), nil
			},
		},
		"formatted": formattedDateType,
		"fromNow":   &fromNowType,
	},
})

var dateType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Date",
	Fields: graphql.Fields{
		"year": &graphql.Field{
			Name: "year",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				d, ok := p.Source.(time.Time)

				if !ok {
					return nil, fmt.Errorf("not a valid date")
				}

				return d.Year(), nil
			},
		},
		"fromNow": &fromNowType,
		"month": &graphql.Field{
			Name: "month",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				d, ok := p.Source.(time.Time)

				if !ok {
					return nil, fmt.Errorf("not a valid date")
				}

				return int(d.Month()), nil
			},
		},
		"day": &graphql.Field{
			Name: "day",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				d, ok := p.Source.(time.Time)

				if !ok {
					return nil, fmt.Errorf("not a valid date")
				}

				return d.Day(), nil
			},
		},
		"formatted": formattedDateType,
	},
})

var timestampInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "TimestampInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"seconds": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
	},
})

var dateInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DateInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"year": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"month": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"day": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
	},
})

var idQueryInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "IDQueryInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"equals": &graphql.InputObjectFieldConfig{
			Type: graphql.ID,
		},
		"oneOf": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.NewNonNull(graphql.ID)),
		},
	},
})

var stringQueryInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "StringQueryInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"equals": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"startsWith": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"endsWith": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"contains": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"oneOf": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.NewNonNull(graphql.String)),
		},
	},
})

var intQueryInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "IntQueryInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"equals": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"lessThan": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"lessThanOrEquals": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"greaterThan": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"greaterThanOrEquals": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var booleanQueryInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "BooleanQueryInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"equals": &graphql.InputObjectFieldConfig{
			Type: graphql.Boolean,
		},
	},
})

var timestampQueryInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "TimestampQueryInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"before": &graphql.InputObjectFieldConfig{
			Type: timestampInputType,
		},
		"after": &graphql.InputObjectFieldConfig{
			Type: timestampInputType,
		},
	},
})

var dateQueryInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DateQueryInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"equals": &graphql.InputObjectFieldConfig{
			Type: dateInputType,
		},
		"before": &graphql.InputObjectFieldConfig{
			Type: dateInputType,
		},
		"onOrBefore": &graphql.InputObjectFieldConfig{
			Type: dateInputType,
		},
		"after": &graphql.InputObjectFieldConfig{
			Type: dateInputType,
		},
		"onOrAfter": &graphql.InputObjectFieldConfig{
			Type: dateInputType,
		},
	},
})

type introsepctionTypeRef struct {
	Name   string                `json:"name"`
	Kind   string                `json:"kind"`
	OfType *introsepctionTypeRef `json:"ofType"`
}

func (t introsepctionTypeRef) String() string {
	if t.Kind == "NON_NULL" {
		return t.OfType.String() + "!"
	}
	if t.Kind == "LIST" {
		return "[" + t.OfType.String() + "]"
	}
	return t.Name
}

type introspectionField struct {
	Args []struct {
		DefaultValue interface{}          `json:"defaultValue"`
		Name         string               `json:"name"`
		Type         introsepctionTypeRef `json:"type"`
	} `json:"args"`
	Name string               `json:"name"`
	Type introsepctionTypeRef `json:"type"`
}

// Represents the result of executing github.com/graphql-go/graphql/testutil.IntrospectionQuery
type IntrospectionQueryResult struct {
	Schema struct {
		MutationType struct {
			Name string `json:"name"`
		} `json:"mutationType"`
		QueryType struct {
			Name string `json:"name"`
		} `json:"queryType"`
		Types []struct {
			EnumValues []struct {
				Name string
			} `json:"enumValues"`
			Fields        []introspectionField `json:"fields"`
			InputFields   []introspectionField `json:"inputFields"`
			Interfaces    interface{}          `json:"interfaces"`
			Kind          string               `json:"kind"`
			Name          string               `json:"name"`
			PossibleTypes interface{}          `json:"possibleTypes"`
		} `json:"types"`
	} `json:"__schema"`
}

// connectionResponse consumes the raw records returned by actions.List() (and similar),
// and wraps them into a Node+Edges structure that is good for the connections pattern
// return type and is expected by the GraphQL schema for the List operation.
// See https://relay.dev/graphql/connections.htm
func connectionResponse(records any, hasNextPage bool) (resp any, err error) {

	recordsList, ok := records.([]map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot cast this: %v to a []map[string]any", records)
	}
	var startCursor string
	var endCursor string
	edges := []map[string]any{}
	for i, record := range recordsList {
		edge := map[string]any{
			"cursor": record["id"],
			"node":   record,
		}
		edges = append(edges, edge)
		if i == 0 {
			startCursor, _ = record["id"].(string)
		}
		if i == len(edges)-1 {
			endCursor, _ = record["id"].(string)
		}
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
