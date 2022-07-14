package actions

import (
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
)

// todo - should this code be in the proto package?

// zeroValueForModel provides a map[string]any that contains all the fields
// that exist in the given proto.Model - with their value set to their
// default, or in (go-speak) zero-values.
func zeroValueForModel(pModel *proto.Model, schema *proto.Schema) (map[string]any, error) {
	zeroValue := map[string]any{}
	var err error
	for _, field := range pModel.Fields {
		if zeroValue[field.Name], err = zeroValueForField(field, schema.Enums); err != nil {
			return nil, err
		}
	}
	return zeroValue, nil
}

// zeroValueForField provides a suitable zero value for the
// given fieldType
func zeroValueForField(field *proto.Field, enums []*proto.Enum) (zeroV any, err error) {
	zeroV = nil

	// todo if the field has a default value use it - and return.
	// nb. defer this because default is not yet accessible from proto.Field

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

func fakeRow(model *proto.Model, enums []*proto.Enum) (row map[string]any, err error) {
	row = map[string]any{}
	for _, field := range model.Fields {
		zeroValue, err := zeroValueForField(field, enums)
		if err != nil {
			return nil, err
		}
		row[field.Name] = zeroValue
	}
	return row, nil
}
