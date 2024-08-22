package tools

import (
	"context"
	"errors"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/rpc/rpc"
)

type Generator struct {
	Schema *proto.Schema
	Tools  []*rpc.ActionConfig
}

var ErrInvalidSchema = errors.New("invalid schema")

// NewGenerator creates a new tooll config generator for the given schema
func NewGenerator(schema *proto.Schema) (*Generator, error) {
	return &Generator{
		Schema: schema,
	}, nil
}

// Generate will generate all the tools for this generator's schema
func (g *Generator) Generate(ctx context.Context) error {
	if g.Schema == nil {
		return ErrInvalidSchema
	}

	// reset any previous tools
	g.Tools = []*rpc.ActionConfig{}

	// generate model tools
	if err := g.generateModelsTools(ctx); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateModelsTools(ctx context.Context) error {
	for _, model := range g.Schema.GetModels() {
		for _, action := range model.GetActions() {
			tool, err := g.actionTool(ctx, model, action)
			if err != nil {
				return err
			}

			g.Tools = append(g.Tools, tool)
		}
	}

	return nil
}

func (g *Generator) actionTool(ctx context.Context, model *proto.Model, action *proto.Action) (*rpc.ActionConfig, error) {
	t := rpc.ActionConfig{
		Name:           "Tool",
		ActionName:     action.Name,
		ActionType:     action.GetType(),
		Implementation: action.GetImplementation(),
	}

	return &t, nil
}
