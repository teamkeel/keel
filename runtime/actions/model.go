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
func zeroValueForModel(pModel *proto.Model) (map[string]any, error) {
	zeroValue := map[string]any{}
	var err error
	for _, field := range pModel.Fields {
		if zeroValue[field.Name], err = zeroValueForField(field.Type.Type); err != nil {
			return nil, err
		}
	}
	return zeroValue, nil
}

// zeroValueForField provides a suitable zero value for the
// given fieldType
func zeroValueForField(fieldType proto.Type) (zeroV any, err error) {
	// todo these are just placeholders to make it compile at the moment, and need much more work.
	switch fieldType {
	case proto.Type_TYPE_STRING:
		return "", nil
	case proto.Type_TYPE_BOOL:
		return false, nil
	case proto.Type_TYPE_INT:
		return 0, nil
	case proto.Type_TYPE_TIMESTAMP:
		return time.Time{}, nil
	case proto.Type_TYPE_DATE:
		return time.Time{}, nil
	case proto.Type_TYPE_ID:
		kid, err := ksuid.NewRandomWithTime(time.Now())
		if err != nil {
			return nil, err
		}
		return kid, nil
	case proto.Type_TYPE_MODEL:
		return "", nil
	case proto.Type_TYPE_CURRENCY:
		return "", nil
	case proto.Type_TYPE_DATETIME:
		return time.Time{}, nil
	case proto.Type_TYPE_ENUM:
		return "", nil
	default:
		return nil, fmt.Errorf("zero value for field type: %s not yet implemented", fieldType)
	}
}

func setFieldsFromInputValues(modelMap map[string]any, args map[string]any) error {
	for paramName, paramValue := range args {
		modelMap[paramName] = paramValue
	}
	return nil
}

func fakeRow(model *proto.Model) (row map[string]any, err error) {
	row = map[string]any{}
	for _, field := range model.Fields {
		zeroValue, err := zeroValueForField(field.Type.Type)
		if err != nil {
			return nil, err
		}
		row[field.Name] = zeroValue
	}
	return row, nil
}
