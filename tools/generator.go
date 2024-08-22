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
	Tools  map[string]*rpc.ActionConfig
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
	g.Tools = map[string]*rpc.ActionConfig{}

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
				Capabilities:   g.capabilities(model, action),
			}

			g.Tools[t.Id] = &t
		}
	}

	return nil
}

// capabilities generates the capabilities/features available for a tool generated for the given action.
// Audit trail is enabled just for GET actions on models that have an UPDATE action.
//
// TODO: Decide on further capabilities
func (g *Generator) capabilities(model *proto.Model, action *proto.Action) *rpc.Capabilities {
	c := &rpc.Capabilities{
		Comments: false,
		Audit:    false,
	}

	// Audit is enabled for get actions for models that also have an update action
	if action.GetType() == proto.ActionType_ACTION_TYPE_GET {
		for _, act := range model.GetActions() {
			if act.GetType() == proto.ActionType_ACTION_TYPE_UPDATE {
				c.Audit = true
				break
			}
		}
	}

	return c
}
