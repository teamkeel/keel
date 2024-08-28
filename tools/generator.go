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
	Config         *rpc.ActionConfig
	Model          *proto.Model
	Action         *proto.Action
	SortableFields []string
}

type Generator struct {
	Schema *proto.Schema
	Tools  map[string]*Tool
}

const fieldNameID = "id"

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
					ApiName:        g.Schema.FindApiName(model.Name, action.Name),
					Name:           casing.ToSentenceCase(action.Name),
					ActionName:     action.Name,
					ActionType:     action.GetType(),
					Implementation: action.GetImplementation(),
					EntitySingle:   strings.ToLower(model.GetName()),
					EntityPlural:   casing.ToPlural(strings.ToLower(model.GetName())),
					Capabilities:   g.makeCapabilities(action),
					Title:          g.makeTitle(action, model),
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

	// decorate further...
	for id, tool := range g.Tools {
		// List actions...
		if tool.Action.IsList() {
			// ... have pagination
			tool.Config.Pagination = &rpc.CursorPaginationConfig{
				Start: &rpc.CursorPaginationConfig_FieldConfig{
					RequestInput:  "after",
					ResponseField: &rpc.JsonPath{Path: "$.pageInfo.startCursor"},
				},
				End: &rpc.CursorPaginationConfig_FieldConfig{
					RequestInput:  "before",
					ResponseField: &rpc.JsonPath{Path: "$.pageInfo.endCursor"},
				},
				PageSize: &rpc.CursorPaginationConfig_PageSizeConfig{
					RequestInput:  "first",
					ResponseField: &rpc.JsonPath{Path: "$.pageInfo.count"},
					DefaultValue:  50,
				},
				NextPage:   &rpc.JsonPath{Path: "$.pageInfo.hasNextPage"},
				TotalCount: &rpc.JsonPath{Path: "$.pageInfo.totalCount"},
			}

			//... and other related actions if applicable (i.e. other list actions defined on the same model)
			// we search for more than one list tool as the results will include the one we're on
			if relatedTools := g.findListTools(tool.Model.Name); len(relatedTools) > 1 {
				for _, relatedID := range relatedTools {
					if id != relatedID {
						tool.Config.RelatedActions = append(tool.Config.RelatedActions, &rpc.ActionLink{
							ToolId: relatedID,
						})
					}
				}
			}
		}

		// get the path of the id response field for this tool
		idResponseFieldPath := tool.getIDResponseFieldPath()

		// entry activity actions for GET and LIST that have an id response
		if idResponseFieldPath != "" && (tool.Action.IsList() || tool.Action.IsGet()) {
			for linkedTool, fieldPath := range g.findAllByIDTools(tool.Model.Name, nil) {
				if linkedTool == id {
					// skip linking to the same tool
					continue
				}

				tool.Config.EntryActivityActions = append(tool.Config.EntryActivityActions, &rpc.ActionLink{
					ToolId: linkedTool,
					Data: []*rpc.DataMapping{
						{
							Key:   fieldPath,
							Value: &rpc.DataMapping_Path{Path: &rpc.JsonPath{Path: idResponseFieldPath}},
						},
					},
				})
			}
		}

		// get entry action for tools that operate on a model instance/s (create/update/list). This is used to link
		if idResponseFieldPath != "" {
			if tool.Action.IsList() || tool.Action.IsUpdate() || tool.Action.Type == proto.ActionType_ACTION_TYPE_CREATE {
				if getToolID := g.findGetByIDTool(tool.Model.Name); getToolID != "" {
					tool.Config.GetEntryAction = &rpc.ActionLink{
						ToolId: getToolID,
						Data: []*rpc.DataMapping{
							{
								Key:   g.Tools[getToolID].getIDInputFieldPath(),
								Value: &rpc.DataMapping_Path{Path: &rpc.JsonPath{Path: idResponseFieldPath}},
							},
						},
					}
				}
			}
		}

		// for all inputs that are IDs that have a get_entry_action link (e.g. used to lookup a related model field),
		// decorate the data mapping now that we have all inputs and responses generated
		for _, input := range tool.Config.Inputs {
			if input.GetEntryAction != nil && input.GetEntryAction.ToolId != "" {
				input.GetEntryAction.Data = []*rpc.DataMapping{
					{
						Key: g.Tools[input.GetEntryAction.ToolId].getIDInputFieldPath(),
						Value: &rpc.DataMapping_Path{
							Path: input.FieldLocation,
						},
					},
				}
			}
		}
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

		// If there are any OrderBy fields, then we find the sortable field names and store them against the tool, to be
		// used later on when generating the response
		if orderBy := msg.GetOrderByField(); orderBy != nil {
			sortableFields := []string{}
			for _, unionMsgName := range orderBy.Type.GetUnionNames() {
				unionMsg := g.Schema.FindMessage(unionMsgName.GetValue())
				if unionMsg == nil {
					return ErrInvalidSchema
				}
				for _, f := range unionMsg.Fields {
					if f.Type.Type == proto.Type_TYPE_SORT_DIRECTION {
						sortableFields = append(sortableFields, f.Name)
					}
				}
			}

			tool.SortableFields = sortableFields
		}
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

			fields, err := g.makeResponsesForMessage(msg, "", tool.SortableFields)
			if err != nil {
				return err
			}
			tool.Config.Response = fields

			continue
		}

		// we don't have a response message, therefore the response will be the model...
		pathPrefix := ""
		// if the action is a list action, we also need to include the pageInfo responses and prefix the results
		if tool.Action.IsList() {
			pathPrefix = ".results[*]"
			tool.Config.Response = append(tool.Config.Response, getPageInfoResponses()...)
		}
		fields, err := g.makeResponsesForModel(tool.Model, pathPrefix, tool.Action.GetResponseEmbeds(), tool.SortableFields)
		if err != nil {
			return err
		}
		tool.Config.Response = append(tool.Config.Response, fields...)
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
			Visible: func() bool {
				return f.IsModelField()
			}(),
		}
		if f.Type.Type == proto.Type_TYPE_ID && f.Type.ModelName != nil {
			// generate action link placeholders
			if lookupToolsIDs := g.findListTools(f.Type.ModelName.Value); len(lookupToolsIDs) > 0 {
				config.LookupAction = &rpc.ActionLink{
					ToolId: lookupToolsIDs[0],
				}
			}

			// create the GetEntry tool link to retrieve the entry for this related model. At this point, not all tools'
			// inputs and repsonses have been generated ; this is a placeholder that will have it's data populated later
			// in the generation process
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

func (g *Generator) makeResponsesForMessage(msg *proto.Message, pathPrefix string, sortableFields []string) ([]*rpc.ResponseFieldConfig, error) {
	fields := []*rpc.ResponseFieldConfig{}

	for _, f := range msg.GetFields() {
		if f.IsMessage() {
			submsg := g.Schema.FindMessage(f.Type.MessageName.Value)
			if submsg == nil {
				return nil, ErrInvalidSchema
			}
			subFields, err := g.makeResponsesForMessage(submsg, "."+f.Name, []string{})
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
			Visible:       true,
			Sortable: func() bool {
				for _, fn := range sortableFields {
					if fn == f.Name {
						return true
					}
				}
				return false
			}(),
		}

		if f.IsFile() {
			config.ImagePreview = true
		}

		fields = append(fields, config)
	}

	return fields, nil
}

// makeResponsesForModel will return an array of response fields for the given model
func (g *Generator) makeResponsesForModel(model *proto.Model, pathPrefix string, embeddings []string, sortableFields []string) ([]*rpc.ResponseFieldConfig, error) {
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
				embeddedFields, err := g.makeResponsesForModel(g.Schema.FindModel(f.ModelName), prefix, fieldEmbeddings, []string{})
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
			Visible:       true,
			Sortable: func() bool {
				for _, fn := range sortableFields {
					if fn == f.Name {
						return true
					}
				}
				return false
			}(),
		}

		if f.IsFile() {
			config.ImagePreview = true
		}

		// if this field is a model, we add a link to the action used to retrieve the related model. Note that inputs are
		// generated first, so we're safe to create a tool/action link now
		if f.IsForeignKey() && f.ForeignKeyInfo.RelatedModelField == fieldNameID {
			if getToolID := g.findGetByIDTool(f.ForeignKeyInfo.RelatedModelName); getToolID != "" {
				config.Link = &rpc.ActionLink{
					ToolId: getToolID,
					Data: []*rpc.DataMapping{
						{
							Key: g.Tools[getToolID].getIDInputFieldPath(),
							Value: &rpc.DataMapping_Path{
								Path: config.FieldLocation,
							},
						},
					},
				}
			}
		}

		fields = append(fields, config)
	}

	return fields, nil
}

// findListTools will search for list tools for the given model
func (g *Generator) findListTools(modelName string) []string {
	ids := []string{}
	for id, tool := range g.Tools {
		if tool.Model.Name == modelName && tool.Action.IsList() {
			ids = append(ids, id)
		}
	}

	return ids
}

// findGetTool will search for a get tool for the given model
func (g *Generator) findGetTool(modelName string) string {
	for id, tool := range g.Tools {
		if tool.Model.Name == modelName && tool.Action.IsGet() {
			return id
		}
	}

	return ""
}

// findGetByIDTool will search for a get tool for the given model that takes in an ID
func (g *Generator) findGetByIDTool(modelName string) string {
	for id, tool := range g.Tools {
		if tool.Model.Name == modelName && tool.Action.IsGet() && tool.hasOnlyIDInput() {
			return id
		}
	}

	return ""
}

// findByIDTools searches for the tools that operate on the given model and take in an ID as an input; optionally
// filtered by an action type. Returns a map of tool IDs and the path of the input field; e.g. getPost: $.id
//
// GET READ DELETE WRITE etc tools are included if they take in only on input (the ID)
// UPDATE tools are included if they take in a where.id input alongside other inputs
func (g *Generator) findAllByIDTools(modelName string, actionType *proto.ActionType) map[string]string {
	toolIds := map[string]string{}
	for id, tool := range g.Tools {
		if actionType != nil && tool.Action.Type != *actionType {
			continue
		}
		if tool.Model.Name != modelName {
			continue
		}

		// if we only have one input, an ID, add and continue
		if tool.hasOnlyIDInput() {
			toolIds[id] = tool.getIDInputFieldPath()
			continue
		}

		// if we have a UPDATE that includes a where.ID
		if tool.Action.IsUpdate() {
			idInputPath := ""
			for _, input := range tool.Config.Inputs {
				if input.FieldType == proto.Type_TYPE_ID && input.FieldLocation.Path == "$.where.id" {
					idInputPath = input.FieldLocation.Path
				}
			}
			if idInputPath != "" {
				toolIds[id] = idInputPath
				continue
			}
		}
	}
	return toolIds
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
	if action.IsGet() {
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
	if action.IsGet() || action.Type == proto.ActionType_ACTION_TYPE_READ {
		fields := model.GetFields()
		if len(fields) > 0 && fields[0].Type.Type == proto.Type_TYPE_STRING {
			return &rpc.StringTemplate{
				Template: "{{." + fields[0].GetName() + "}}",
			}
		}
	}

	return nil
}

// getPageInfoResponses will return the responses for pageInfo (by default available on all autogenerated LIST actions)
func getPageInfoResponses() []*rpc.ResponseFieldConfig {
	return []*rpc.ResponseFieldConfig{
		{
			FieldLocation: &rpc.JsonPath{Path: "$.pageInfo.count"},
			FieldType:     proto.Type_TYPE_INT,
			DisplayName:   "Count",
			Visible:       false,
		},
		{
			FieldLocation: &rpc.JsonPath{Path: "$.pageInfo.totalCount"},
			FieldType:     proto.Type_TYPE_INT,
			DisplayName:   "Total count",
			Visible:       false,
		},
		{
			FieldLocation: &rpc.JsonPath{Path: "$.pageInfo.hasNextPage"},
			FieldType:     proto.Type_TYPE_BOOL,
			DisplayName:   "Has next page",
			Visible:       false,
		},
		{
			FieldLocation: &rpc.JsonPath{Path: "$.pageInfo.startCursor"},
			FieldType:     proto.Type_TYPE_STRING,
			DisplayName:   "Start cursor",
			Visible:       false,
		},
		{
			FieldLocation: &rpc.JsonPath{Path: "$.pageInfo.endCursor"},
			FieldType:     proto.Type_TYPE_STRING,
			DisplayName:   "End cursor",
			Visible:       false,
		},
	}
}

// hasOnlyIDInput checks if the tool takes only one input, an ID
func (t *Tool) hasOnlyIDInput() bool {
	if len(t.Config.Inputs) != 1 {
		return false
	}
	for _, input := range t.Config.Inputs {
		if input.FieldType != proto.Type_TYPE_ID {
			return false
		}
	}

	return true
}

// getIDInputFieldPath returns the path of the first input field that's an ID
func (t *Tool) getIDInputFieldPath() string {
	for _, input := range t.Config.Inputs {
		if input.FieldType == proto.Type_TYPE_ID && input.DisplayName == casing.ToSentenceCase(fieldNameID) {
			return input.FieldLocation.Path
		}
	}

	return ""
}

// getIDResponseFieldPath returns the path of the first response field that's an ID at top level (i.e. results[*].id
// rather than results[*].embedded.id for list actions). Returns empty string if ID is not part of the response
func (t *Tool) getIDResponseFieldPath() string {
	expectedPath := "$.id"
	if t.Action.IsList() {
		expectedPath = "$.results[*].id"
	}
	for _, response := range t.Config.Response {
		if response.FieldType == proto.Type_TYPE_ID && response.FieldLocation.Path == expectedPath {
			return response.FieldLocation.Path
		}
	}

	return ""
}
