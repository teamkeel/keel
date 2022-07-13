package proto

import (
	"fmt"

	"github.com/samber/lo"
)

// ModelNames provides a (sorted) list of all the Model names used in the
// given schema.
func ModelNames(p *Schema) []string {
	return sortedStrings(lo.Map(p.Models, func(x *Model, _ int) string {
		return x.Name
	}))
}

// FieldNames provides a (sorted) list of the fields in the model of
// the given name.
func FieldNames(m *Model) []string {
	return lo.Map(m.Fields, func(x *Field, _ int) string {
		return x.Name
	})
}

// ModelsExists returns true if the given schema contains a
// model with the given name.
func ModelExists(models []*Model, name string) bool {
	for _, m := range models {
		if m.Name == name {
			return true
		}
	}
	return false
}

// FindModel locates the model of the given name.
// It panics if there is no model of that name.
func FindModel(models []*Model, name string) *Model {
	model, _, found := lo.FindIndexOf(models, func(m *Model) bool {
		return m.Name == name
	})
	if !found {
		panic(fmt.Sprintf("There is no model with the name: %s", name))
	}
	return model
}

// FindOperation locates the operation of the given name in the given model.
// It panics if there is no model of that name.
func FindOperation(models *Model, name string) *Operation {
	op, _, found := lo.FindIndexOf(models.Operations, func(op *Operation) bool {
		return op.Name == name
	})
	if !found {
		panic(fmt.Sprintf("There is no operation with the name: %s", name))
	}
	return op
}

// FindModels locates and returns the models whose names match up with those
// specified in the given names to find.
func FindModels(allModels []*Model, namesToFind []string) (foundModels []*Model) {
	for _, candidateModel := range allModels {
		if lo.Contains(namesToFind, candidateModel.Name) {
			foundModels = append(foundModels, candidateModel)
		}
	}
	return foundModels
}

func FindField(models []*Model, modelName string, fieldName string) *Field {
	model := FindModel(models, modelName)
	for _, field := range model.Fields {
		if field.Name == fieldName {
			return field
		}
	}
	panic(fmt.Sprintf("There is no <%s> field in model: %s", fieldName, modelName))
}
