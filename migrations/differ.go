package migrations

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/protoqry"
)

// ProtoDeltas provides information about the differences between the two
// given schemas.
func ProtoDeltas(old, new *proto.Schema) (*Differences, error) {
	diffs := NewDifferences()

	// Models added or removed.
	diffs.ModelsRemoved, diffs.ModelsAdded = lo.Difference(
		protoqry.ModelNames(old),
		protoqry.ModelNames(new))

	// Fields added or removed
	for _, m := range new.Models {
		modelName := m.Name
		newNames := protoqry.FieldNames(protoqry.FindModel(new.Models, modelName))
		oldNames := []string{}
		if protoqry.ModelExists(old.Models, modelName) {
			oldNames = protoqry.FieldNames(protoqry.FindModel(old.Models, modelName))
		}
		diffs.FieldsRemoved[modelName], diffs.FieldsAdded[modelName] = lo.Difference(oldNames, newNames)
	}

	return diffs, nil
}
