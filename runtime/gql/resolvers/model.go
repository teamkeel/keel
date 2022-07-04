package resolvers

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
)

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

// A zeroValue is similar to go's concept of a zero value, except that it offers not
// only a "zero" value, but also an "illustrative" value.
type zeroValue struct {
	zero         any
	illustrative any
}

// zeroValueForField provides a suitable zero value (and an illustrative value) for the
// given fieldType
func zeroValueForField(fieldType proto.Type) (zeroValue, error) {
	// todo these are just placeholders to make it compile at the moment, and need much more work.
	switch fieldType {
	case proto.Type_TYPE_STRING:
		return zeroValue{"", "my-string"}, nil
	case proto.Type_TYPE_BOOL:
		return zeroValue{false, false}, nil
	case proto.Type_TYPE_INT:
		return zeroValue{0, 1234}, nil
	case proto.Type_TYPE_TIMESTAMP:
		return zeroValue{"", "2014-11-12T11:45:26.371Z"}, nil
	case proto.Type_TYPE_DATE:
		return zeroValue{"", "2014-11-12"}, nil
	case proto.Type_TYPE_ID:
		return zeroValue{"", "e7a6d74c-8a08-49b5-85ae-d9816f1f15ca"}, nil
	case proto.Type_TYPE_MODEL:
		return zeroValue{"", "Person"}, nil
	case proto.Type_TYPE_CURRENCY:
		return zeroValue{"", "EUR"}, nil
	case proto.Type_TYPE_DATETIME:
		return zeroValue{"", "2014-11-12T11:45:26.371Z"}, nil
	case proto.Type_TYPE_ENUM:
		return zeroValue{"unknownEnum", "XLARGE"}, nil
	case proto.Type_TYPE_IDENTITY:
		return zeroValue{"", "foo@bar.com"}, nil
	case proto.Type_TYPE_IMAGE:
		return zeroValue{"", "/some-images/flower.jpg"}, nil

	default:
		return zeroValue{}, fmt.Errorf("zero value for field type: %s not yet implemented", fieldType)
	}
}

func setFieldsFromInputValues(modelMap map[string]any, resolveParams graphql.ResolveParams) error {
	for paramName, paramValue := range resolveParams.Args {
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
		row[field.Name] = zeroValue.illustrative
	}
	return row, nil
}
