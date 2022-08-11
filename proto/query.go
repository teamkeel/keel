package proto

import (
	"sort"

	"github.com/samber/lo"
)

// ModelNames provides a (sorted) list of all the Model names used in the
// given schema.
func ModelNames(p *Schema) []string {
	names := lo.Map(p.Models, func(x *Model, _ int) string {
		return x.Name
	})
	sort.Strings(names)
	return names
}

// FieldNames provides a (sorted) list of the fields in the model of
// the given name.
func FieldNames(m *Model) []string {
	names := lo.Map(m.Fields, func(x *Field, _ int) string {
		return x.Name
	})
	sort.Strings(names)
	return names
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
func FindModel(models []*Model, name string) *Model {
	model, _ := lo.Find(models, func(m *Model) bool {
		return m.Name == name
	})
	return model
}

func FilterOperations(p *Schema, filter func(op *Operation) bool) (ops []*Operation) {
	for _, model := range p.Models {
		operations := model.Operations

		for _, o := range operations {
			if filter(o) {
				ops = append(ops, o)
			}
		}
	}

	return ops
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
	return nil
}

// ModelHasField returns true IFF the schema contains a model of the given name AND
// that model has a field of the given name.
func ModelHasField(schema *Schema, model string, field string) bool {
	for _, m := range schema.Models {
		if m.Name != model {
			continue
		}
		for _, f := range m.Fields {
			if f.Name == field {
				return true
			}
		}
	}
	return false
}

// OperationHasInput returns true if the given Operation defines
// an input of the given name.
func OperationHasInput(op *Operation, name string) bool {
	for _, input := range op.Inputs {
		if input.Name == name {
			return true
		}
	}
	return false
}
