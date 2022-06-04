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
	old *proto.Schema
	new *proto.Schema
}

func NewProtoDiffer(old, new *proto.Schema) *ProtoDiffer {
	return &ProtoDiffer{
		old: old,
		new: new,
	}
}

func (d *ProtoDiffer) Analyse() (diffs Differences, err error) {
	diffs.ModelsRemoved, diffs.ModelsAdded = lo.Difference(
		protoqry.AllModelNames(d.old),
		protoqry.AllModelNames(d.new))

	return diffs, nil
}
