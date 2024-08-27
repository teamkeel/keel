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

type Tool struct {
	Config *rpc.ActionConfig
	Model  *proto.Model
	Action *proto.Action
}

type Generator struct {
	Schema *proto.Schema
	Tools  map[string]*Tool
}

var ErrInvalidSchema = errors.New("invalid schema")

// NewGenerator creates a new tooll config generator for the given schema
func NewGenerator(schema *proto.Schema) (*Generator, error) {
	return &Generator{
		Schema: schema,
	}, nil
}

// GetConfigs will return the action configs that have been generated
func (g *Generator) GetConfigs() []*rpc.ActionConfig {
	cfgs := []*rpc.ActionConfig{}
	for _, t := range g.Tools {
		cfgs = append(cfgs, t.Config)
	}

	return cfgs
}

// Generate will generate all the tools for this generator's schema
func (g *Generator) Generate(ctx context.Context) error {
	if g.Schema == nil {
		return ErrInvalidSchema
	}

	// reset any previous tools
	g.Tools = map[string]*Tool{}

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
			t := Tool{
				Config: &rpc.ActionConfig{
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

					// Pagination:           nil,

					// RelatedActions:       nil,
					// EntryActivityActions: nil,
					// EmbeddedActions:      nil,
					// GetEntryAction:       nil,
				},
				Model:  model,
				Action: action,
			}

			g.Tools[t.Config.Id] = &t
		}
	}
}

func (g *Generator) decorateTools() error {
	if err := g.generateInputs(); err != nil {
		return fmt.Errorf("generating inputs: %w", err)
	}

	if err := g.generateResponses(); err != nil {
		return fmt.Errorf("generating responses: %w", err)
	}

	return nil
}

// generateInputs will make the inputs for all tools
func (g *Generator) generateInputs() error {
	for _, tool := range g.Tools {
		// if the action does not have a input message, it means we don't have any inputs for this tool
		if tool.Action.InputMessageName == "" {
			continue
		}

		// get the input message
		msg := g.Schema.FindMessage(tool.Action.InputMessageName)
		if msg == nil {
			return ErrInvalidSchema
		}

		fields, err := g.makeInputsForMessage(msg, "")
		if err != nil {
			return err
		}
		tool.Config.Inputs = fields
	}

	return nil
}

// generateResponses will make the responses for all tools
func (g *Generator) generateResponses() error {
	for _, tool := range g.Tools {
		// if the action has a response message, let's generate it
		if tool.Action.ResponseMessageName != "" {
			// get the message message
			msg := g.Schema.FindMessage(tool.Action.ResponseMessageName)
			if msg == nil {
				return ErrInvalidSchema
			}

			fields, err := g.makeResponsesForMessage(msg, "")
			if err != nil {
				return err
			}
			tool.Config.Response = fields

			continue
		}

		// we don't have a response message, therefore the response will be the model
		pathPrefix := ""
		if tool.Action.Type == proto.ActionType_ACTION_TYPE_LIST {
			pathPrefix = ".results[*]"
		}
		fields, err := g.makeResponsesForModel(tool.Model, pathPrefix, tool.Action.GetResponseEmbeds())
		if err != nil {
			return err
		}
		tool.Config.Response = fields
	}

	return nil
}

func (g *Generator) makeInputsForMessage(msg *proto.Message, pathPrefix string) ([]*rpc.RequestFieldConfig, error) {
	fields := []*rpc.RequestFieldConfig{}

	for _, f := range msg.GetFields() {
		if f.IsMessage() {
			submsg := g.Schema.FindMessage(f.Type.MessageName.Value)
			if submsg == nil {
				return nil, ErrInvalidSchema

			}
			subFields, err := g.makeInputsForMessage(submsg, "."+f.Name)
			if err != nil {
				return nil, err
			}
			fields = append(fields, subFields...)

			continue
		}

		config := &rpc.RequestFieldConfig{
			FieldLocation: &rpc.JsonPath{Path: `$` + pathPrefix + "." + f.Name},
			FieldType:     f.Type.Type,
			DisplayName:   casing.ToSentenceCase(f.Name),
		}
		if f.Type.Type == proto.Type_TYPE_ID && f.Type.ModelName != nil {
			// generate action link placeholders
			if lookupToolID := g.findListTool(f.Type.ModelName.Value); lookupToolID != "" {
				config.LookupAction = &rpc.ActionLink{
					ToolId: lookupToolID,
				}
			}

			if entryToolID := g.findGetTool(f.Type.ModelName.Value); entryToolID != "" {
				config.GetEntryAction = &rpc.ActionLink{
					ToolId: entryToolID,
				}
			}
		}

		fields = append(fields, config)
	}

	return fields, nil
}

func (g *Generator) makeResponsesForMessage(msg *proto.Message, pathPrefix string) ([]*rpc.ResponseFieldConfig, error) {
	fields := []*rpc.ResponseFieldConfig{}

	for _, f := range msg.GetFields() {
		if f.IsMessage() {
			submsg := g.Schema.FindMessage(f.Type.MessageName.Value)
			if submsg == nil {
				return nil, ErrInvalidSchema
			}
			subFields, err := g.makeResponsesForMessage(submsg, "."+f.Name)
			if err != nil {
				return nil, err
			}
			fields = append(fields, subFields...)

			continue
		}

		config := &rpc.ResponseFieldConfig{
			FieldLocation: &rpc.JsonPath{Path: `$` + pathPrefix + "." + f.Name},
			FieldType:     f.Type.Type,
			DisplayName:   casing.ToSentenceCase(f.Name),
		}

		if f.IsFile() {
			config.ImagePreview = true
		}

		fields = append(fields, config)
	}

	return fields, nil
}

// makeResponsesForModel will return an array of response fields for the given model
func (g *Generator) makeResponsesForModel(model *proto.Model, pathPrefix string, embeddings []string) ([]*rpc.ResponseFieldConfig, error) {
	fields := []*rpc.ResponseFieldConfig{}

	for _, f := range model.GetFields() {
		if f.IsTypeModel() {
			// models are only included if they are embedded
			found := false
			fieldEmbeddings := []string{}

			for _, embed := range embeddings {
				frags := strings.Split(embed, ".")
				if frags[0] == f.Name {
					found = true
					// if we have to embed a child model for this field, we need to pass them through with the first segment removed
					if len(frags) > 1 {
						fieldEmbeddings = append(fieldEmbeddings, strings.Join(frags[1:], "."))
					}
				}
			}
			if found {
				prefix := pathPrefix + "." + f.Name
				if f.IsHasMany() {
					prefix = prefix + "[*]"
				}
				embeddedFields, err := g.makeResponsesForModel(g.Schema.FindModel(f.ModelName), prefix, fieldEmbeddings)
				if err != nil {
					return nil, err
				}
				fields = append(fields, embeddedFields...)
			}

			continue
		}

		config := &rpc.ResponseFieldConfig{
			FieldLocation: &rpc.JsonPath{Path: `$` + pathPrefix + "." + f.Name},
			FieldType:     f.Type.Type,
			DisplayName:   casing.ToSentenceCase(f.Name),
		}

		if f.IsFile() {
			config.ImagePreview = true
		}

		fields = append(fields, config)
	}

	return fields, nil
}

// findListTool will search for a list tool for the given model
func (g *Generator) findListTool(modelName string) string {
	for id, tool := range g.Tools {
		if tool.Model.Name == modelName && tool.Action.Type == proto.ActionType_ACTION_TYPE_LIST {
			return id
		}
	}

	return ""
}

// findListTool will search for a get tool for the given model that takes in an ID
func (g *Generator) findGetTool(modelName string) string {
	for id, tool := range g.Tools {
		if tool.Model.Name == modelName && tool.Action.Type == proto.ActionType_ACTION_TYPE_GET {
			return id
		}
	}

	return ""
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
