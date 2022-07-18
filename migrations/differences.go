package migrations

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

// Differences encapsulates the differences between two proto.Proto objects,
// for the purposes of informing database migrations.
type Differences struct {
	ModelsAdded   []string
	ModelsRemoved []string

	// FieldsAdded refers to models that exist in both the old and new schemas, but which
	// have been newly introduced in the new schema. The map is keyed on model names.
	FieldsAdded map[string][]string

	// FieldsRemoved refers to models that exist in both the old and new schemas, but which
	// have been removed in the new schema. The map is keyed on model names.
	FieldsRemoved map[string][]string
}

// NewDifferences provides information about the differences between the two
// given schemas.
func NewDifferences(old, new *proto.Schema) *Differences {
	diffs := &Differences{
		FieldsAdded:   map[string][]string{},
		FieldsRemoved: map[string][]string{},
	}

	oldModels := proto.ModelNames(old)
	newModels := proto.ModelNames(new)
	modelsInCommon := modelsPresentInBothOldAndNew(old, new)

	// Models added or removed.
	diffs.ModelsRemoved, diffs.ModelsAdded = lo.Difference(oldModels, newModels)

	// Fields added or removed
	for _, modelName := range modelsInCommon {
		oldFieldNames := proto.FieldNames(proto.FindModel(old.Models, modelName))
		newFieldNames := proto.FieldNames(proto.FindModel(new.Models, modelName))
		diffs.FieldsRemoved[modelName], diffs.FieldsAdded[modelName] = lo.Difference(oldFieldNames, newFieldNames)
	}

	return diffs
}

func modelsPresentInBothOldAndNew(old, new *proto.Schema) []string {
	oldNames := proto.ModelNames(old)
	newNames := proto.ModelNames(new)
	namesInCommon := lo.Intersect(oldNames, newNames)
	return namesInCommon
}
