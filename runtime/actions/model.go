package actions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/expressions"
)

// initialValueForModel provides a map[string]any that corresponds to all the fields
// that exist in the given proto.Model - with their value set to schema-default value if
// one is specified for this field in the schema, or failing that our built-in default value for
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
// given field. It first tries to use its schema-default.
// After that it tries to use our built-in default for the field's type.
// But it can also return nil without an error when it cannot do either of those things.
func initialValueForField(field *proto.Field, enums []*proto.Enum) (zeroV any, err error) {
	zeroV = nil

	switch {
	case field.DefaultValue != nil && field.DefaultValue.Expression != nil:
		{
			v, err := schemaDefault(field) // Will need more arguments / context later.
			if err != nil {
				return nil, err
			}
			return v, nil
		}
	case field.DefaultValue != nil && field.DefaultValue.UseZeroValue:
		v, err := builtinDefault(field, enums)
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

func builtinDefault(field *proto.Field, enums []*proto.Enum) (any, error) {
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

	default:
		// When we cannot provide a built-in default - we assing nil as the value
		// to use.
		return nil, nil
	}
}

func schemaDefault(field *proto.Field) (any, error) {
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
		return toNative(v, field.Type.Type), nil
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

// toMap provides two values in potentially different forms for the given proto.OperationInput.
// The first is good to insert into the corresponding DB column. The second is good return from
// the top level Action functions like Create.
func toMap(in any, inputType proto.Type) (forDB any, toReturn any, err error) {
	switch inputType {

	// Start with some special cases that require some intervention.

	case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:

		// The input is seconds-since Jan 1st 1970
		seconds, ok := in.(int64)
		if !ok {
			return nil, nil, fmt.Errorf("cannot cast %+v to int64", in)
		}
		tm := time.Unix(seconds, 0)
		return tm.Format(time.RFC3339), tm, nil

	case proto.Type_TYPE_DATE:
		// The input is of the form 18/03/2011
		s, ok := in.(string)
		if !ok {
			return nil, nil, fmt.Errorf("cannot cast %+v to string", in)
		}
		segs := strings.Split(s, `/`)
		day, _ := strconv.Atoi(segs[0])
		month, _ := strconv.Atoi(segs[1])
		year, _ := strconv.Atoi(segs[2])
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		return date.Format(time.UnixDate), date, nil

	// The general case is to return the input unchanged.
	default:
		return in, in, nil
	}
}
