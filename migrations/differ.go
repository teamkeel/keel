package migrations

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/protoqry"
)

// A ProtoDiffer knows how to measure those of the differences between two
// proto.Schema objects - that are sufficient to govern a database migration
// from one to the other. For example models present in one and not the other,
// or a field that exists for a certain model in both, but which has differing
// constraints or type from one to the other.
type ProtoDiffer struct {
	previousSchema *proto.Schema
	incomingSchema *proto.Schema
}

func NewProtoDiffer(old, new *proto.Schema) *ProtoDiffer {
	return &ProtoDiffer{
		previousSchema: old,
		incomingSchema: new,
	}
}

// Analyse provides information about the differences between the two
// schemas given at construction time.
func (d *ProtoDiffer) Analyse() (*Differences, error) {
	diffs := NewDifferences()

	// Models added or removed.
	diffs.ModelsRemoved, diffs.ModelsAdded = lo.Difference(
		protoqry.ModelNames(d.previousSchema),
		protoqry.ModelNames(d.incomingSchema))

	// For models that exist in both schemas, which fields are newly introduced in
	// the new one, and which have been dropped in the new one?
	for _, m := range d.incomingSchema.Models {
		modelName := m.Name
		newNames := protoqry.FieldNames(protoqry.FindModel(d.incomingSchema.Models, modelName))
		oldNames := []string{}
		if protoqry.ModelExists(d.previousSchema.Models, modelName) {
			oldNames = protoqry.FieldNames(protoqry.FindModel(d.previousSchema.Models, modelName))
		}
		diffs.FieldsRemoved[modelName], diffs.FieldsAdded[modelName] = lo.Difference(oldNames, newNames)
	}

	return diffs, nil
}
