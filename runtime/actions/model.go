package actions

import (
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/expressions"
)

// initialValueForModel provides a map[string]any that corresponds to all the fields
// that exist in the given proto.Model - with their value set to their default value if
// one is specified for this field in the schema, or failing that our built-in "zero value" for
// the corresponding type. E.g. emtpy string for string fields, or integer zero for a Number field.
func initialValueForModel(pModel *proto.Model, schema *proto.Schema) (map[string]any, error) {
	zeroValue := map[string]any{}
	var err error
	for _, field := range pModel.Fields {
		if zeroValue[field.Name], err = initialValueForField(field, schema.Enums); err != nil {
			return nil, err
		}
	}
	return zeroValue, nil
}

// initialValueForField provides a suitable initial value for the
// given field. It first tries to use the default value specified in the schema for this field.
// After that it tries to use our built-in or zero value for the field's type.
// But it can also return nil without an error when it cannot do either of those things.
func initialValueForField(field *proto.Field, enums []*proto.Enum) (zeroV any, err error) {
	zeroV = nil

	switch {
	case field.DefaultValue != nil && field.DefaultValue.Expression != nil:
		{
			v, err := evalDefaultValueExpression(field) // Will need more arguments / context later.
			if err != nil {
				return nil, err
			}
			return v, nil
		}
	case field.DefaultValue != nil && field.DefaultValue.UseZeroValue:
		v, err := evalZeroValue(field, enums)
		if err != nil {
			return nil, err
		}
		return v, nil
	default:
		// We cannot provide a sensibly initialized value, but that is not an error, because the
		// value will likely be provided later from the operation input values.
		return nil, nil
	}
}

func evalZeroValue(field *proto.Field, enums []*proto.Enum) (any, error) {
	fType := field.Type.Type
	rpt := field.Type.Repeated
	now := time.Now()
	kid, err := ksuid.NewRandomWithTime(now)
	if err != nil {
		return nil, err
	}

	switch {

	case fType == proto.Type_TYPE_STRING && !rpt:
		return "", nil
	case fType == proto.Type_TYPE_STRING && rpt:
		return []string{}, nil

	case fType == proto.Type_TYPE_BOOL && !rpt:
		return false, nil
	case fType == proto.Type_TYPE_BOOL && rpt:
		return []bool{}, nil

	case fType == proto.Type_TYPE_INT && !rpt:
		return 0, nil
	case fType == proto.Type_TYPE_INT && rpt:
		return []int{}, nil

	case fType == proto.Type_TYPE_TIMESTAMP && !rpt:
		return now, nil
	case fType == proto.Type_TYPE_TIMESTAMP && rpt:
		return []time.Time{}, nil

	case fType == proto.Type_TYPE_DATE && !rpt:
		return now, nil
	case fType == proto.Type_TYPE_DATE && rpt:
		return []time.Time{}, nil

	case fType == proto.Type_TYPE_ID && !rpt:
		return kid, nil
	case fType == proto.Type_TYPE_ID && rpt:
		return []ksuid.KSUID{}, nil

	case fType == proto.Type_TYPE_MODEL && !rpt:
		return "", nil
	case fType == proto.Type_TYPE_MODEL && rpt:
		return []string{}, nil

	case fType == proto.Type_TYPE_CURRENCY && !rpt:
		return "", nil
	case fType == proto.Type_TYPE_CURRENCY && rpt:
		return []string{}, nil

	case fType == proto.Type_TYPE_DATETIME && !rpt:
		return now, nil
	case fType == proto.Type_TYPE_DATETIME && rpt:
		return []time.Time{}, nil

	case fType == proto.Type_TYPE_ENUM && !rpt:
		return defaultForEnum(field, enums), nil
	case fType == proto.Type_TYPE_ENUM && rpt:
		return []string{}, nil

	default:
		return nil, fmt.Errorf("zero value for field: %s not yet implemented", fType)
	}
}

func defaultForEnum(field *proto.Field, enums []*proto.Enum) string {
	for _, enum := range enums {
		if field.Type.EnumName.Value == enum.Name {
			return enum.Values[0].Name
		}
	}
	return ""
}

func setFieldsFromInputValues(modelMap map[string]any, args map[string]any) error {
	for paramName, paramValue := range args {
		modelMap[paramName] = paramValue
	}
	return nil
}

func evalDefaultValueExpression(field *proto.Field) (any, error) {
	source := field.DefaultValue.Expression.Source
	expr, err := expressions.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("cannot Parse this expression: %s", source)
	}
	switch {
	case expressions.IsValue(expr):
		v, err := expressions.ToValue(expr)
		if err != nil {
			return nil, err
		}
		return toNative(v), nil
	default:
		return nil, fmt.Errorf("expressions that are not simple values are not yet supported")
	}
}

func fakeRow(model *proto.Model, enums []*proto.Enum) (row map[string]any, err error) {
	row = map[string]any{}
	for _, field := range model.Fields {
		zeroValue, err := initialValueForField(field, enums)
		if err != nil {
			return nil, err
		}
		row[field.Name] = zeroValue
	}
	return row, nil
}
