package graphql

import (
	"errors"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/sirupsen/logrus"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/schema/parser"
)

func getInput(schema *proto.Schema, operation *proto.Operation, args map[string]any) any {
	inputAsMap, ok := args["input"].(map[string]any)
	if !ok {
		inputAsMap = map[string]any{}
	}

	switch operation.Type {
	case proto.OperationType_OPERATION_TYPE_READ, proto.OperationType_OPERATION_TYPE_WRITE:
		inputMessage := proto.FindMessage(schema.Messages, operation.InputMessageName)

		if inputMessage.Name == parser.MessageFieldTypeAny {
			// we can't do any more processing of an Any type
			return args["input"]
		}

		// we have a message type that we want to parse
		inputAsMap = parseTypes(inputMessage, operation, inputAsMap)
	case proto.OperationType_OPERATION_TYPE_GET, proto.OperationType_OPERATION_TYPE_CREATE, proto.OperationType_OPERATION_TYPE_DELETE:
		inputMessage := proto.FindMessage(schema.Messages, operation.InputMessageName)
		inputAsMap = parseTypes(inputMessage, operation, inputAsMap)
	case proto.OperationType_OPERATION_TYPE_UPDATE, proto.OperationType_OPERATION_TYPE_LIST:
		if where, ok := inputAsMap["where"].(map[string]any); ok {
			whereMessage := proto.FindWhereInputMessage(schema, operation.Name)
			if whereMessage != nil {
				inputAsMap["where"] = parseTypes(whereMessage, operation, where)
			}
		}
		if values, ok := inputAsMap["values"].(map[string]any); ok {
			valuesMessage := proto.FindValuesInputMessage(schema, operation.Name)
			if valuesMessage != nil {
				inputAsMap["values"] = parseTypes(valuesMessage, operation, values)
			}
		}
	}

	return inputAsMap
}

func ActionFunc(schema *proto.Schema, operation *proto.Operation) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		scope := actions.NewScope(p.Context, operation, schema)

		input := getInput(schema, operation, p.Args)

		res, headers, err := actions.Execute(scope, input)

		if err != nil {
			var runtimeErr common.RuntimeError
			if !errors.As(err, &runtimeErr) {
				logrus.Error(err)
				err = common.RuntimeError{
					Code:    common.ErrInternal,
					Message: "error executing request",
				}
			} else {
				logrus.Trace(err)
			}
			return nil, err
		}

		rootValue := p.Info.RootValue.(map[string]interface{})
		headersValue := rootValue["headers"].(map[string][]string)
		for k, v := range headers {
			headersValue[k] = v
		}

		if operation.Type == proto.OperationType_OPERATION_TYPE_LIST {
			// actions.Execute() returns any but a list action will return a map
			m, _ := res.(map[string]any)
			return connectionResponse(m)
		}

		return res, nil
	}
}

func parseTypes(message *proto.Message, operation *proto.Operation, values map[string]any) map[string]any {
	for k, v := range values {
		field, found := lo.Find(message.Fields, func(in *proto.MessageField) bool {
			return in.Name == k
		})

		if !found {
			continue
		}

		if operation.Type == proto.OperationType_OPERATION_TYPE_LIST && field.IsModelField() {
			if field.Type.Type == proto.Type_TYPE_MESSAGE && field.Type.MessageName.Value == "DateQuery_input" {
				listOpMap := v.(map[string]any)

				for kListOp, vListOp := range listOpMap {
					listOpMap[kListOp] = convertDate(vListOp)
				}
				values[k] = listOpMap
			}
			if field.Type.Type == proto.Type_TYPE_MESSAGE && field.Type.MessageName.Value == "TimestampQuery_input" {
				listOpMap := v.(map[string]any)
				for kListOp, vListOp := range listOpMap {
					listOpMap[kListOp] = convertTimestamp(vListOp)
				}
				values[k] = listOpMap
			}
		} else {
			if field.Type.Type == proto.Type_TYPE_DATE {
				values[k] = convertDate(v)
			}
			if field.Type.Type == proto.Type_TYPE_DATETIME {
				values[k] = convertTimestamp(v)
			}
		}
	}

	return values
}

func convertDate(value any) time.Time {
	dateMap, ok := value.(map[string]any)
	if !ok {
		panic("date must be a map")
	}

	day, okDay := dateMap["day"].(int)
	month, okMonth := dateMap["month"].(int)
	year, okYear := dateMap["year"].(int)

	if !(okDay && okMonth && okYear) {
		panic("date badly formatted")
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func convertTimestamp(value any) time.Time {
	timeMap, ok := value.(map[string]any)
	if !ok {
		panic("date must be a map")
	}
	seconds, ok := timeMap["seconds"].(int) // todo: should be int64
	if !ok {
		panic("time badly formatted")
	}

	return time.Unix(int64(seconds), 0).UTC()
}
