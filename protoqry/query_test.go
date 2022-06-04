package protoqry

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestAllModelNames(t *testing.T) {
	p := &proto.Schema{
		Models: []*proto.Model{
			{
				Name: "A",
			},
			{
				Name: "B",
			},
		},
	}
	result := AllModelNames(p)
	require.Equal(t, []string{"A", "B"}, result)
}
