package tools

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/rpc/rpc"
)

// GenerateTools will return a map of tool configurations generated for the given schema, keyed by their ids
func GenerateTools(ctx context.Context, schema *proto.Schema) (map[string]*rpc.ActionConfig, error) {
	if schema == nil {
		return nil, nil
	}

	gen, err := NewGenerator(schema)
	if err != nil {
		return nil, fmt.Errorf("creating tool generator: %w", err)
	}

	if err := gen.Generate(ctx); err != nil {
		return nil, fmt.Errorf("generating tools: %w", err)
	}

	return gen.Tools, nil
}
