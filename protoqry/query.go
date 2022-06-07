// package protoqry provides a variety of query functions about
// what is in a proto.Schema. For example - a list of all the
// model names.
package protoqry

import (
	"github.com/samber/lo"

	"github.com/teamkeel/keel/proto"
)

// ModelNames provides a (sorted) list of all the Model names used in the
// given schema.
func ModelNames(p *proto.Schema) []string {
	return sortedStrings(lo.Map(p.Models, func(x *proto.Model, _ int) string {
		return x.Name
	}))
}

// FieldNames provides a (sorted) list of the fields in the model of
// the given name.
func FieldNames(m *proto.Model) []string {
	return lo.Map(m.Fields, func(x *proto.Field, _ int) string {
		return x.Name
	})
}

// ModelsExists returns true if the given schema contains a
// model with the given name.
// todo move this family into proto package
func ModelExists(models []*proto.Model, name string) bool {
	_, _, found := lo.FindIndexOf(models, func(m *proto.Model) bool {
		return m.Name == name
	})
	return lo.Ternary(found, true, false)
}

// FindModel locates the model of the given name.
// It panics if there is no model of that name.
func FindModel(models []*proto.Model, name string) *proto.Model {
	model, _, found := lo.FindIndexOf(models, func(m *proto.Model) bool {
		return m.Name == name
	})
	if !found {
		panic("There is no model of that name")
	}
	return model
}
