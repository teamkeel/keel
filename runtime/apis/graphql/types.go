package graphql

import (
	"fmt"
	"time"

	"github.com/bykof/gostradamus"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/nleeper/goment"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// anyType represents a flexible scalar type that can be literally anything that is valid in JSON
// e.g string, number, boolean, list of objects, list of numbers, list of strings etc
// Primarily used for arbitrary functions when an input/response is defined with the type 'Any'
var anyType = graphql.NewScalar(graphql.ScalarConfig{
	Name: parser.MessageFieldTypeAny,
	ParseValue: func(value interface{}) interface{} {
		return value
	},
	Serialize: func(value interface{}) interface{} {
		return value
	},
	// ParseLiteral is used to parse literal values in a graphql query/mutation
	// this applies to values hardcoded in a graphql query in graphiql playground too
	ParseLiteral: func(valueAST ast.Value) interface{} {
		// ast.Value is GraphQL's internal representation of a value of many types
		// it isn't so simple to parse the actual underlying value to Go, so we need to do that in parseASTValue
		return parseASTValue(valueAST)
	},
	Description: parser.MessageFieldTypeAny,
})

// parseASTValue attempts to parse the contents of the graphql value AST which represents many types of values (strings, bools, lists etc)
func parseASTValue(v ast.Value) interface{} {
	// todo: inspect the type switch here to ensure we covered all of the standard types
	switch underlying := v.(type) {
	case *ast.StringValue:
		return underlying.Value
	case *ast.ListValue:
		return lo.Map(underlying.Values, func(v ast.Value, _ int) interface{} {
			return parseASTValue(v)
		})
	case *ast.BooleanValue:
		return underlying.Value
	case *ast.EnumValue:
		return underlying.Value
	case *ast.IntValue:
		return underlying.Value
	case *ast.ObjectValue:
		return lo.Reduce(underlying.Fields, func(acc map[string]any, cur *ast.ObjectField, _ int) map[string]any {
			acc[cur.Name.Value] = parseASTValue(cur.Value)
			return acc
		}, map[string]any{})
	case *ast.FloatValue:
		return underlying.Value
	default:
		// best guess attempt at grabbing the underlying value from the graphql ast node
		// note: imperfect
		return underlying.GetValue()
	}
}

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
		"iso8601": &graphql.Field{
			Name:        "iso8601",
			Description: "ISO8601 representation of the timestamp",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				t, err := sourceToTime(p.Source)

				if err != nil {
					return nil, err
				}

				// iso8601 layout
				return t.Format("2006-01-02T15:04:05.000Z"), nil
			},
		},
		"formatted": formattedDateType,
		"fromNow":   &fromNowType,
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

var dateType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Date",
	Fields: graphql.Fields{
		"iso8601": &graphql.Field{
			Name:        "iso8601",
			Description: "ISO8601 representation of the date",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				t, err := sourceToTime(p.Source)

				if err != nil {
					return nil, err
				}

				// iso8601 layout
				return t.Format("2006-01-02T15:04:05.000Z"), nil
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
	proto.Type_TYPE_ANY:      anyType,
}

var timestampInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "TimestampInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"iso8601": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var dateInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DateInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"iso8601": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
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
	proto.Type_TYPE_ANY:       anyType,
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
