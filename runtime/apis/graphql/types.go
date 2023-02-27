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
			Resolve: func(p graphql.ResolveParams) (any, error) {
				return p.Source, nil
			},
			Type: graphql.NewNonNull(graphql.Boolean),
		},
	},
})

var pageInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PageInfo",
	Fields: graphql.Fields{
		"hasNextPage": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether there are results after the current page.",
		},
		"startCursor": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The ID cursor of the first node on the current page.",
		},
		"endCursor": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The ID cursor of the last node on the current page.",
		},
		"totalCount": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Total count of nodes on the current page.",
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
		t, err := sourceToTime(p.Source)

		if err != nil {
			return nil, err
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
		dateTime := gostradamus.DateTimeFromTime(*t)

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
				t, err := sourceToTime(p.Source)

				if err != nil {
					return nil, err
				}

				return t.Unix(), nil
			},
		},
		"year": &graphql.Field{
			Name: "year",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				t, err := sourceToTime(p.Source)

				if err != nil {
					return nil, err
				}

				return t.Year(), nil
			},
		},
		"month": &graphql.Field{
			Name: "month",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				t, err := sourceToTime(p.Source)

				if err != nil {
					return nil, err
				}

				return int(t.Month()), nil
			},
		},
		"day": &graphql.Field{
			Name: "day",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				t, err := sourceToTime(p.Source)

				if err != nil {
					return nil, err
				}

				return t.Day(), nil
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
				t, err := sourceToTime(p.Source)

				if err != nil {
					return nil, err
				}

				return t.Year(), nil
			},
		},
		"month": &graphql.Field{
			Name: "month",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				t, err := sourceToTime(p.Source)

				if err != nil {
					return nil, err
				}

				return int(t.Month()), nil
			},
		},
		"day": &graphql.Field{
			Name: "day",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				t, err := sourceToTime(p.Source)

				if err != nil {
					return nil, err
				}

				return t.Day(), nil
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

// for fields where the underlying source is a date/datetime
// the actual underlying field value may either be a time.Time
// or an ISO8601 string. So this method handles differing inputs for the
// source value, and returns a time.Time
func sourceToTime(source interface{}) (*time.Time, error) {
	switch v := source.(type) {
	case time.Time:
		return &v, nil
	case string:
		t, err := time.Parse(time.RFC3339, v)

		if err != nil {
			return nil, err
		}

		return &t, nil
	default:
		return nil, fmt.Errorf("%v not a valid date / time", source)
	}
}
