package migrations

import "github.com/teamkeel/keel/proto"

// A ProtoDiffer knows how to measure the differences between two
// proto.Schema objects - for the purposes of generating database
// migrations.
type ProtoDiffer struct {
}

func NewProtoDiffer() *ProtoDiffer {
	return &ProtoDiffer{}
}

func (d *ProtoDiffer) Analyse(incumbent, incoming *proto.Schema) (diffs *Differences, err error) {
	return nil, nil
}
