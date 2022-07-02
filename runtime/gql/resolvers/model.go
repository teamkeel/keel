package resolvers

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
)

// A ModelMap represents the values of the fields in a Keel model, in the form of a primitive
// go datatype, that is good for returning from graphql resolvers.
type ModelMap map[string]any

// zeroValueForModel provides a ModelMap that contains all the fields
// that exist in the given proto.Model - with their value set to their
// default, or in (go-speak) zero-values.
func zeroValueForModel(pModel *proto.Model) (ModelMap, error) {
	mMap := ModelMap{}
	var err error
	for _, field := range pModel.Fields {
		if mMap[field.Name], err = zeroValueForField(field.Type.Type); err != nil {
			return nil, err
		}
	}
	return mMap, nil
}

// zeroValueForField provides a suitable zero value for the
// given fieldType.
func zeroValueForField(fieldType proto.Type) (any, error) {
	// todo these are just placeholders to make it compile at the moment, and need much more work.
	switch fieldType {
	case proto.Type_TYPE_STRING:
		return "", nil
	case proto.Type_TYPE_BOOL:
		return false, nil
	case proto.Type_TYPE_INT:
		return 0, nil
	case proto.Type_TYPE_TIMESTAMP:
		return "", nil
	case proto.Type_TYPE_DATE:
		return "", nil
	case proto.Type_TYPE_ID:
		return graphql.ID, nil
	case proto.Type_TYPE_MODEL:
		return "", nil
	case proto.Type_TYPE_CURRENCY:
		return "", nil
	case proto.Type_TYPE_DATETIME:
		return "", nil
	case proto.Type_TYPE_ENUM:
		return "", nil
	case proto.Type_TYPE_IDENTITY:
		return "", nil
	case proto.Type_TYPE_IMAGE:
		return "", nil

	default:
		return nil, fmt.Errorf("zero value for field type: %s not yet implemented", fieldType)
	}
}

func (mm ModelMap) setFieldsFromInputValues(resolveParams graphql.ResolveParams) error {
	for paramName, paramValue := range resolveParams.Args {
		mm[paramName] = paramValue
	}
	return nil
}
