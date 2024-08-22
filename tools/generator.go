package tools

import (
	"context"
	"errors"
	"strings"

	"github.com/teamkeel/keel/casing"
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

	// first pass at generating tools;
	if err := g.scaffoldTools(); err != nil {
		return err
	}

	return nil
}

// scaffoldTools will generate all the basic tools. These will be incomplete configurations, with fields and
// relations between them not yet filled in
//
// For each model's actions, we will scaffold the `ActionConfig`s. These will not yet contain all request fields,
// response fields and any related/embedded tools, as these need to reference each other, so we first scaffold them and
// the completed generation is done later on
func (g *Generator) scaffoldTools() error {
	for _, model := range g.Schema.GetModels() {
		for _, action := range model.GetActions() {
			t := rpc.ActionConfig{
				Id:             action.Name,
				Name:           casing.ToSentenceCase(action.Name),
				ActionName:     action.Name,
				ActionType:     action.GetType(),
				Implementation: action.GetImplementation(),
				EntitySingle:   strings.ToLower(model.GetName()),
				EntityPlural:   casing.ToPlural(strings.ToLower(model.GetName())),
			}
			g.Tools = append(g.Tools, &t)
		}
	}

	return nil
}
