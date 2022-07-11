package actions

import (
	"context"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"gorm.io/gorm"
)

func Create(ctx context.Context, model *proto.Model, args map[string]any) (map[string]any, error) {
	// We'll populate a map[string]any to represent the resolved model field values, and
	// use that map, to write a record into the database, and as the return value.
	modelMap, err := zeroValueForModel(model)
	if err != nil {
		return nil, err
	}
	if err = setFieldsFromInputValues(modelMap, args); err != nil {
		return nil, err
	}

	err = createRecordInDatabase(runtimectx.GetDB(ctx), model, modelMap)
	if err != nil {
		return nil, err
	}

	return modelMap, nil
}

func createRecordInDatabase(db *gorm.DB, model *proto.Model, modelMap map[string]any) error {
	// this is where we write a record to the database.
	q := db.Table(model.Name)
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
	return nil
}
