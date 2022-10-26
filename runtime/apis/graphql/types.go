package graphql

import (
	"fmt"
	"time"

	"github.com/bykof/gostradamus"
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
)

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

var protoTypeToGraphQLOutput = map[proto.Type]graphql.Output{
	proto.Type_TYPE_ID:       graphql.ID,
	proto.Type_TYPE_STRING:   graphql.String,
	proto.Type_TYPE_INT:      graphql.Int,
	proto.Type_TYPE_BOOL:     graphql.Boolean,
	proto.Type_TYPE_DATETIME: timestampType,
	proto.Type_TYPE_DATE:     dateType,
	proto.Type_TYPE_SECRET:   graphql.String,
}

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

var protoTypeToGraphQLInput = map[proto.Type]graphql.Input{
	proto.Type_TYPE_ID:        graphql.ID,
	proto.Type_TYPE_STRING:    graphql.String,
	proto.Type_TYPE_INT:       graphql.Int,
	proto.Type_TYPE_BOOL:      graphql.Boolean,
	proto.Type_TYPE_TIMESTAMP: timestampInputType,
	proto.Type_TYPE_DATETIME:  timestampInputType,
	proto.Type_TYPE_DATE:      dateInputType,
	proto.Type_TYPE_SECRET:    graphql.String,
	proto.Type_TYPE_PASSWORD:  graphql.String,
}

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

var protoTypeToGraphQLQueryInput = map[proto.Type]graphql.Input{
	proto.Type_TYPE_ID:        idQueryInputType,
	proto.Type_TYPE_STRING:    stringQueryInputType,
	proto.Type_TYPE_INT:       intQueryInputType,
	proto.Type_TYPE_BOOL:      booleanQueryInput,
	proto.Type_TYPE_TIMESTAMP: timestampQueryInputType,
	proto.Type_TYPE_DATETIME:  timestampQueryInputType,
	proto.Type_TYPE_DATE:      dateQueryInputType,
}
