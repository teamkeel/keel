package migrations

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

// ProtoDeltas provides information about the differences between the two
// given schemas.
func ProtoDeltas(old, new *proto.Schema) (*Differences, error) {
	diffs := NewDifferences()

	oldModels := proto.ModelNames(old)
	newModels := proto.ModelNames(new)
	modelsIncommon := modelsPresentInBothOldAndNew(old, new)

	// Models added or removed.
	diffs.ModelsRemoved, diffs.ModelsAdded = lo.Difference(oldModels, newModels)

	// Fields added or removed
	for _, modelName := range modelsIncommon {
		oldFieldNames := proto.FieldNames(proto.FindModel(old.Models, modelName))
		newFieldNames := proto.FieldNames(proto.FindModel(new.Models, modelName))
		diffs.FieldsRemoved[modelName], diffs.FieldsAdded[modelName] = lo.Difference(oldFieldNames, newFieldNames)
	}

	return diffs, nil
}

func modelsPresentInBothOldAndNew(old, new *proto.Schema) []string {
	oldNames := proto.ModelNames(old)
	newNames := proto.ModelNames(new)
	namesInCommon := lo.Intersect(oldNames, newNames)
	return namesInCommon
}
