package tools

import (
	"context"
	"errors"
	"fmt"
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
	g.scaffoldTools()

	// then decorate the tools with all the relevant options
	if err := g.decorateTools(); err != nil {
		return fmt.Errorf("decorating tools: %w", err)
	}

	return nil
}

// scaffoldTools will generate all the basic tools. These will be incomplete configurations, with fields and
// relations between them not yet filled in
//
// For each model's actions, we will scaffold the `ActionConfig`s. These will not yet contain all request fields,
// response fields and any related/embedded tools, as these need to reference each other, so we first scaffold them and
// the completed generation is done later on
func (g *Generator) scaffoldTools() {
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
				Capabilities:   g.makeCapabilities(action),
				Title:          g.makeTitle(action, model),
				// ApiName: nil,
				// Inputs:               nil,
				// Response:             nil,
				// RelatedActions:       nil,
				// Pagination:           nil,
				// EntryActivityActions: nil,
				// EmbeddedActions:      nil,
				// GetEntryAction:       nil,
			}

			g.Tools[t.Id] = &t
		}
	}
}

func (g *Generator) decorateTools() error {
	for id := range g.Tools {
		inputs, err := g.generateInputs(id)
		if err != nil {
			return fmt.Errorf("generating inputs: %w", err)
		}
		g.Tools[id].Inputs = inputs

		// switch tool.ActionType {
		// case proto.ActionType_ACTION_TYPE_CREATE:
		// case proto.ActionType_ACTION_TYPE_GET:
		// case proto.ActionType_ACTION_TYPE_LIST:
		// case proto.ActionType_ACTION_TYPE_UPDATE:
		// case proto.ActionType_ACTION_TYPE_DELETE:
		// case proto.ActionType_ACTION_TYPE_READ:
		// case proto.ActionType_ACTION_TYPE_WRITE:
		// }
	}

	return nil
}

// generateInputs will make the inputs for the given tool
func (g *Generator) generateInputs(toolID string) ([]*rpc.RequestFieldConfig, error) {
	tool := g.Tools[toolID]
	action := g.Schema.FindAction(tool.ActionName)
	if action == nil {
		return nil, ErrInvalidSchema
	}

	// if the action does not have a input message, it means we don't have any inputs for this tool
	if action.InputMessageName == "" {
		return []*rpc.RequestFieldConfig{}, nil
	}

	// get the input message
	msg := g.Schema.FindMessage(action.InputMessageName)
	if msg == nil {
		return nil, ErrInvalidSchema
	}

	fields := []*rpc.RequestFieldConfig{}
	// TODO: implement this

	// for _, f := range msg.GetFields() {
	// 	fields = append(fields, &rpc.RequestFieldConfig{
	// 		FieldLocation: &rpc.JsonPath{Path: f.Name},
	// 		FieldType:     f.Type.Type,
	// 	})
	// }

	return fields, nil
}

// makeCapabilities generates the makeCapabilities/features available for a tool generated for the given action.
// Audit trail is enabled just for GET actions
// Comments are enabled just for GET actions
func (g *Generator) makeCapabilities(action *proto.Action) *rpc.Capabilities {
	c := &rpc.Capabilities{
		Comments: false,
		Audit:    false,
	}

	// Audit is enabled for get actions for models that also have an update action
	if action.GetType() == proto.ActionType_ACTION_TYPE_GET {
		c.Audit = true
		c.Comments = true
	}

	return c
}

// makeTitle will create a string template to be used as a title for a tool.
//
// For GET/READ actions:
//   - The title will be a template including the value of the first field of the model, only if that field is a text field
func (g *Generator) makeTitle(action *proto.Action, model *proto.Model) *rpc.StringTemplate {
	if action.Type == proto.ActionType_ACTION_TYPE_GET || action.Type == proto.ActionType_ACTION_TYPE_READ {
		fields := model.GetFields()
		if len(fields) > 0 && fields[0].Type.Type == proto.Type_TYPE_STRING {
			return &rpc.StringTemplate{
				Template: "{{." + fields[0].GetName() + "}}",
			}
		}
	}

	return nil
}
