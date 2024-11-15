package actions

import (
	"fmt"
	"strconv"
	"time"

	"github.com/relvacode/iso8601"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/types"
)

// TransformInputs will traverse through the input data structure and will ensure that values are correctly typed.
// This is necessary because we need the correct types when generating to SQL and because the JSON and RPC APIs
// don't type correctly when parsing the input JSON (for example, "Number" values become floats).
func TransformInputs(schema *proto.Schema, message *proto.Message, input map[string]any, isFunction bool) (map[string]any, error) {
	input, err := transform(schema, message, input, isFunction)
	if err != nil {
		return nil, err
	}

	return input, nil
}

func transform(schema *proto.Schema, message *proto.Message, input map[string]any, forFunctions bool) (map[string]any, error) {
	var err error

	for _, f := range message.Fields {
		if v, has := input[f.Name]; has {
			switch f.Type.Type {
			case proto.Type_TYPE_MESSAGE:
				nested := schema.FindMessage(f.Type.MessageName.Value)

				if f.Type.Repeated {
					arr := v.([]any)
					for i, el := range arr {
						arr[i], err = transform(schema, nested, el.(map[string]any), forFunctions)
						if err != nil {
							return nil, err
						}
					}
					input[f.Name] = arr
				} else {
					if v == nil {
						input[f.Name] = nil
					} else {
						input[f.Name], err = transform(schema, nested, v.(map[string]any), forFunctions)
						if err != nil {
							return nil, err
						}
					}
				}

			case proto.Type_TYPE_INT:
				input[f.Name], err = parseItem(v, f.Type.Repeated, toInt)
			case proto.Type_TYPE_DECIMAL:
				input[f.Name], err = parseItem(v, f.Type.Repeated, toFloat)
			case proto.Type_TYPE_BOOL:
				input[f.Name], err = parseItem(v, f.Type.Repeated, toBool)
			case proto.Type_TYPE_DATE:
				input[f.Name], err = parseItem(v, f.Type.Repeated, toDate)
			case proto.Type_TYPE_TIMESTAMP, proto.Type_TYPE_DATETIME:
				input[f.Name], err = parseItem(v, f.Type.Repeated, toTimestamp)
			case proto.Type_TYPE_VECTOR:
				input[f.Name], err = parseItem(v, true, toFloat)
			case proto.Type_TYPE_UNION, proto.Type_TYPE_ANY, proto.Type_TYPE_MODEL, proto.Type_TYPE_OBJECT:
				return input, nil
			case proto.Type_TYPE_TIME_PERIOD:
				input[f.Name], err = parseItem(v, f.Type.Repeated, toTimePeriod)
			case proto.Type_TYPE_FILE:
				if forFunctions {
					input[f.Name], err = parseItem(v, f.Type.Repeated, toInlineFileForFunctions)
				} else {
					input[f.Name], err = parseItem(v, f.Type.Repeated, toString)
				}
			default:
				input[f.Name], err = parseItem(v, f.Type.Repeated, toString)
			}

			if err != nil {
				return nil, err
			}
		}
	}

	return input, nil
}

func parseItem[T any](v any, isArray bool, parse func(any) (T, error)) (any, error) {
	var err error

	if v == nil {
		return nil, nil
	}

	if isArray {
		arr := v.([]any)
		values := make([]T, len(arr))

		for i, t := range arr {
			values[i], err = parse(t)
			if err != nil {
				return nil, err
			}
		}
		return values, nil
	} else {
		return parse(v)
	}
}

var toBool = func(value any) (bool, error) {
	switch t := value.(type) {
	case bool:
		return t, nil
	case string:
		return strconv.ParseBool(t)
	default:
		return false, fmt.Errorf("incompatible type %T parsing to bool", t)
	}
}

var toString = func(value any) (string, error) {
	switch t := value.(type) {
	case string:
		return t, nil
	default:
		return "", fmt.Errorf("incompatible type %T parsing to string", t)
	}
}

var toInt = func(value any) (int, error) {
	switch t := value.(type) {
	case int:
		return t, nil
	case float32:
		return int(t), nil
	case float64:
		return int(t), nil
	case string:
		return strconv.Atoi(t)
	default:
		return 0, fmt.Errorf("incompatible type %T parsing to int", t)
	}
}

var toFloat = func(value any) (float64, error) {
	switch t := value.(type) {
	case int:
		return float64(t), nil
	case float32:
		return float64(t), nil
	case float64:
		return t, nil
	default:
		return 0, fmt.Errorf("incompatible type %T parsing to float", t)
	}
}

var toTimestamp = func(value any) (types.Timestamp, error) {
	switch t := value.(type) {
	case string:
		parsed, err := iso8601.ParseString(t)
		return types.Timestamp{Time: parsed}, err
	case time.Time:
		return types.Timestamp{Time: t}, nil
	default:
		return types.Timestamp{}, fmt.Errorf("incompatible type %T parsing to Timestamp", t)
	}
}

var toDate = func(value any) (types.Date, error) {
	switch t := value.(type) {
	case string:
		parsed, err := iso8601.ParseString(t)
		return types.Date{Time: parsed}, err
	case time.Time:
		return types.Date{Time: t}, nil
	default:
		return types.Date{}, fmt.Errorf("incompatible type %T parsing to Date", t)
	}
}

var toInlineFileForFunctions = func(value any) (map[string]any, error) {
	switch t := value.(type) {
	case string:
		return map[string]any{
			"__typename": "InlineFile",
			"dataURL":    t,
		}, nil
	default:
		return nil, fmt.Errorf("incompatible type %T parsing to inline file for functions", t)
	}
}

var toTimePeriod = func(value any) (types.TimePeriod, error) {
	switch t := value.(type) {
	case map[string]interface{}:
		p, ok := t["period"].(string)
		if !ok {
			return types.TimePeriod{}, fmt.Errorf("incompatible period for time period: %v", t["period"])
		}
		o, ok := t["offset"].(float64)
		if !ok {
			return types.TimePeriod{}, fmt.Errorf("incompatible offset for time period: %v", t["offset"])
		}
		c, ok := t["complete"].(bool)
		if !ok {
			return types.TimePeriod{}, fmt.Errorf("incompatible complete for time period: %v", t["complete"])
		}
		return types.TimePeriod{
			Period:   p,
			Offset:   int(o),
			Complete: c,
		}, nil
	default:
		return types.TimePeriod{}, fmt.Errorf("incompatible type %T parsing to time period", t)
	}
}
