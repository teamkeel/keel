package resolvers

import (
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
)

// A CreateOperationResolver provides a Resolve method that matches the signature needed for
// a graphql.FieldResolveFn. It is implemented as an object that holds state, so that
// individual instances can be constructed with access to the data they need when their
// Resolve method gets called.
type CreateOperationResolver struct {
	op    *proto.Operation
	model *proto.Model
}

func NewCreateOperationResolver(op *proto.Operation, model *proto.Model) *CreateOperationResolver {
	return &CreateOperationResolver{
		op:    op,
		model: model,
	}
}

func (r *CreateOperationResolver) Resolve(p graphql.ResolveParams) (any, error) {

	// We'll populate a map[string]any to represent the resolved model field values, and
	// use that map, to write a record into the database, and as the return value
	// from the resolver.

	modelMap, err := zeroValueForModel(r.model)
	if err != nil {
		return nil, err
	}
	if err = setFieldsFromInputValues(modelMap, p); err != nil {
		return nil, err
	}

	fieldValuesFromDB, err := createRecordInDatabase(r.model, modelMap)
	if err != nil {
		return nil, err
	}
	for fieldName, fieldValue := range fieldValuesFromDB {
		modelMap[fieldName] = fieldValue
	}

	return modelMap, nil
}

func createRecordInDatabase(model *proto.Model, modelMap map[string]any) (valuesFromDb map[string]any, err error) {
	// this is where we write a record to the database.

	// and return the field values that are created as a side effect.

	// We'll just give it an "id" value as an illustration for now.
	return map[string]any{
		"id": uuid.NewString(),
	}, nil
}
