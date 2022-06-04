package protoqry

import (
	"github.com/samber/lo"

	"github.com/teamkeel/keel/proto"
)

// AllModelNames provides a list of all the Model names used in the
// given schema - sorted alphanumerically.
func AllModelNames(p *proto.Schema) []string {
	return sortedStrings(lo.Map(p.Models, func(x *proto.Model, _ int) string {
		return x.Name
	}))
}
