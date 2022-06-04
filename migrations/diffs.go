package migrations

// Differences encapsulates the differences between two proto.Proto objects,
// for the purposes of informing database migrations.
type Differences struct {
	ModelsAdded   []string
	ModelsRemoved []string
}
