package resolvers

import (
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"gorm.io/gorm"
)

// A CreateOperationResolver provides a Resolve method that matches the signature needed for
// a graphql.FieldResolveFn. It is implemented as an object that holds state, so that
// individual instances can be constructed with access to the data they need when their
// Resolve method gets called.
type CreateOperationResolver struct {
	op    *proto.Operation
	model *proto.Model
	db    *gorm.DB
}

func NewCreateOperationResolver(db *gorm.DB, op *proto.Operation, model *proto.Model) *CreateOperationResolver {
	return &CreateOperationResolver{
		db:    db,
		op:    op,
		model: model,
	}
}

// func Create(ctx context.Context, args map[string]any) (map[string]any, error) {
// 	// do stuff
// }

// runtime
//   graphql
//     actions.create()
//   rpc
//     actions.create()
//   actions

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

	fieldValuesFromDB, err := r.createRecordInDatabase(r.model, modelMap)
	if err != nil {
		return nil, err
	}
	for fieldName, fieldValue := range fieldValuesFromDB {
		modelMap[fieldName] = fieldValue
	}

	return modelMap, nil
}

func (r CreateOperationResolver) createRecordInDatabase(model *proto.Model, modelMap map[string]any) (valuesFromDb map[string]any, err error) {
	// this is where we write a record to the database.
	q := r.db.Table(model.Name)
	_ = q

	/*
		foo, bar := r.db.Insert()
		foo, bar := q.Insert()

			q := db.Table(inflection.Plural(strcase.ToSnake(model.Name)))

						selects := []string{}
						for _, field := range model.Fields {
							selects = append(selects, strcase.ToSnake(field.Name))
						}
	*/

	// and return the field values that are created as a side effect.

	// We'll just give it an "id" value as an illustration for now.
	return map[string]any{
		"id": uuid.NewString(),
	}, nil
}
