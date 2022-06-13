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

	oldModels := protoqry.ModelNames(old)
	newModels := protoqry.ModelNames(new)
	modelsIncommon := modelsPresentInBothOldAndNew(old, new)

	// Models added or removed.
	diffs.ModelsRemoved, diffs.ModelsAdded = lo.Difference(oldModels, newModels)

	// Fields added or removed
	for _, modelName := range modelsIncommon {
		oldFieldNames := protoqry.FieldNames(protoqry.FindModel(old.Models, modelName))
		newFieldNames := protoqry.FieldNames(protoqry.FindModel(new.Models, modelName))
		diffs.FieldsRemoved[modelName], diffs.FieldsAdded[modelName] = lo.Difference(oldFieldNames, newFieldNames)
	}

	return diffs, nil
}

func modelsPresentInBothOldAndNew(old, new *proto.Schema) []string {
	oldNames := protoqry.ModelNames(old)
	newNames := protoqry.ModelNames(new)
	namesInCommon := lo.Intersect(oldNames, newNames)
	return namesInCommon
}
