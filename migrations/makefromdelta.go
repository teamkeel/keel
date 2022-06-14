package migrations

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
)

func MakeMigrationsFromSchemaDifference(oldProto, newProto *proto.Schema) (theSQL string, err error) {
	differences, err := ProtoDeltas(oldProto, newProto)
	if err != nil {
		return "", fmt.Errorf("could not analyse differences: %v", err)
	}
	// Are there any new tables?
	for _, newModelName := range differences.ModelsAdded {
		model := proto.FindModel(newProto.Models, newModelName)
		theSQL += "\n"
		theSQL += createTable(model)
	}

	// Have any tables disappeared?
	for _, droppedModel := range differences.ModelsRemoved {
		theSQL += "\n"
		theSQL += dropTable(droppedModel)
	}

	// Have any fields been added to models that are present in both old and new schema?
	for modelName, fieldsAdded := range differences.FieldsAdded {
		for _, fieldName := range fieldsAdded {
			field := proto.FindField(newProto.Models, modelName, fieldName)
			theSQL += "\n"
			theSQL += createField(modelName, field)
		}
	}

	// Have any fields been removed from models that are present in both old and new schema?
	for modelName, fieldsRemoved := range differences.FieldsRemoved {
		for _, fieldName := range fieldsRemoved {
			field := proto.FindField(oldProto.Models, modelName, fieldName)
			theSQL += "\n"
			theSQL += dropField(modelName, field.Name)
		}
	}

	return theSQL, nil
}
