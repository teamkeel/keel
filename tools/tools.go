package tools

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

// GenerateTools will return a map of tool configurations generated for the given schema
func GenerateTools(ctx context.Context, schema *proto.Schema) ([]*toolsproto.Tool, error) {
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

	tools := []*toolsproto.Tool{}
	for _, cfg := range gen.GetConfigs() {
		tools = append(tools, &toolsproto.Tool{
			Config: cfg,
			Slug:   strcase.ToKebab(cfg.ActionName),
			Id:     strcase.ToKebab(cfg.ActionName),
		})
	}

	return tools, nil
}
