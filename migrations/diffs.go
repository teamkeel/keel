package migrations

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

func NewDifferences() *Differences {
	return &Differences{
		FieldsAdded:   map[string][]string{},
		FieldsRemoved: map[string][]string{},
	}
}
