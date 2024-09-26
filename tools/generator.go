package tools

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

type Tool struct {
	Config         *toolsproto.ActionConfig
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

// NewGenerator creates a new tool config generator for the given schema
func NewGenerator(schema *proto.Schema) (*Generator, error) {
	return &Generator{
		Schema: schema,
	}, nil
}

// GetConfigs will return the action configs that have been generated, in alphabetical order
func (g *Generator) GetConfigs() []*toolsproto.ActionConfig {
	cfgs := []*toolsproto.ActionConfig{}
	ids := []string{}
	for id := range g.Tools {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		cfgs = append(cfgs, g.Tools[id].Config)
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
				Config: &toolsproto.ActionConfig{
					Id:             action.GetName(),
					ApiNames:       g.Schema.FindApiNames(model.Name, action.Name),
					Name:           casing.ToSentenceCase(action.Name),
					ActionName:     action.GetName(),
					ModelName:      model.GetName(),
					ActionType:     action.GetType(),
					Implementation: action.GetImplementation(),
					EntitySingle:   strings.ToLower(casing.ToSentenceCase(model.GetName())),
					EntityPlural:   casing.ToPlural(strings.ToLower(casing.ToSentenceCase(model.GetName()))),
					Capabilities:   g.makeCapabilities(action),
					Title:          g.makeTitle(action, model),
				},
				Model:  model,
				Action: action,
			}

			// List actions have pagination
			if action.IsList() {
				t.Config.Pagination = &toolsproto.CursorPaginationConfig{
					Start: &toolsproto.CursorPaginationConfig_FieldConfig{
						RequestInput:  "after",
						ResponseField: &toolsproto.JsonPath{Path: "$.pageInfo.startCursor"},
					},
					End: &toolsproto.CursorPaginationConfig_FieldConfig{
						RequestInput:  "before",
						ResponseField: &toolsproto.JsonPath{Path: "$.pageInfo.endCursor"},
					},
					PageSize: &toolsproto.CursorPaginationConfig_PageSizeConfig{
						RequestInput:  "first",
						ResponseField: &toolsproto.JsonPath{Path: "$.pageInfo.count"},
						DefaultValue:  50,
					},
					NextPage:   &toolsproto.JsonPath{Path: "$.pageInfo.hasNextPage"},
					TotalCount: &toolsproto.JsonPath{Path: "$.pageInfo.totalCount"},
				}
			}

			g.Tools[t.Config.ActionName] = &t
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

	g.generateRelatedActionsLinks()
	g.generateEntryActivityActionsLinks()
	g.generateGetEntryActionLinks()
	g.generateEmbeddedActionLinks()
	g.generateCreateEntryActionLinks()

	// decorate further...
	for _, tool := range g.Tools {
		// for all inputs that are IDs that have a get_entry_action link (e.g. used to lookup a related model field),
		// decorate the data mapping now that we have all inputs and responses generated
		for _, input := range tool.Config.Inputs {
			if input.GetEntryAction != nil && input.GetEntryAction.ToolId != "" {
				input.GetEntryAction.Data = []*toolsproto.DataMapping{
					{
						Key:  g.Tools[input.GetEntryAction.ToolId].getIDInputFieldPath(),
						Path: input.FieldLocation,
					},
				}
			}
		}
	}

	return nil
}

// generateRelatedActionsLinks will traverse the tools and generate the RelatedActions links:
//   - For LIST actions = other list actions for the same model
func (g *Generator) generateRelatedActionsLinks() {
	for id, tool := range g.Tools {
		if !tool.Action.IsList() {
			continue
		}

		// we search for more than one list tool as the results will include the one we're on
		if relatedTools := g.findListTools(tool.Model.Name); len(relatedTools) > 1 {
			for _, relatedID := range relatedTools {
				if id != relatedID {
					tool.Config.RelatedActions = append(tool.Config.RelatedActions, &toolsproto.ActionLink{
						ToolId: relatedID,
					})
				}
			}
		}
	}
}

// generateEntryActivityActionsLinks will traverse the tools and generate the EntryActivityActions links:
//   - For LIST/GET actions that have a model ID response = other actions on the same model that take an id as an input
func (g *Generator) generateEntryActivityActionsLinks() {
	for id, tool := range g.Tools {
		// get the path of the id response field for this tool
		idResponseFieldPath := tool.getIDResponseFieldPath()
		// skip if we don't have an id response field or the tool is not List or Get
		if idResponseFieldPath == "" || (!tool.Action.IsList() && !tool.Action.IsGet()) {
			continue
		}

		// entry activity actions for GET and LIST that have an id response

		// TODO: once go is upgraded to 1.21+, refactor to sort via maps package
		inputPaths := g.findAllByIDTools(tool.Model.Name, id)
		toolIds := []string{}
		for id := range inputPaths {
			toolIds = append(toolIds, id)
		}
		sort.Strings(toolIds)
		for _, toolID := range toolIds {
			tool.Config.EntryActivityActions = append(tool.Config.EntryActivityActions, &toolsproto.ActionLink{
				ToolId: toolID,
				Data: []*toolsproto.DataMapping{
					{
						Key:  inputPaths[toolID],
						Path: &toolsproto.JsonPath{Path: idResponseFieldPath},
					},
				},
			})
		}
	}
}

// generateGetEntryActionLinks will traverse the tools and generate the GetEntryAction links:
//   - For LIST/UPDATE/CREATE = a GET action used to retrieve the model by id
func (g *Generator) generateGetEntryActionLinks() {
	for _, tool := range g.Tools {
		// get the path of the id response field for this tool
		idResponseFieldPath := tool.getIDResponseFieldPath()
		if idResponseFieldPath == "" {
			continue
		}
		// get entry action for tools that operate on a model instance/s (create/update/list).
		if tool.Action.IsList() || tool.Action.IsUpdate() || tool.Action.Type == proto.ActionType_ACTION_TYPE_CREATE {
			if getToolID := g.findGetByIDTool(tool.Model.Name); getToolID != "" {
				tool.Config.GetEntryAction = &toolsproto.ActionLink{
					ToolId: getToolID,
					Data: []*toolsproto.DataMapping{
						{
							Key:  g.Tools[getToolID].getIDInputFieldPath(),
							Path: &toolsproto.JsonPath{Path: idResponseFieldPath},
						},
					},
				}
			}
		}
	}
}

// generateCreateEntryActionLinks will traverse the tools and generate the CreateEntryAction links:
//   - Applicable to LIST/GET actions: a CREATE action used to make a model of the same type
func (g *Generator) generateCreateEntryActionLinks() {
	for _, tool := range g.Tools {
		if tool.Action.IsList() || tool.Action.IsGet() {
			if createToolId := g.findCreateTool(tool.Model.Name); createToolId != "" {
				//TODO: improvement: add datamapping from list actions to the create action if there are any filtered fields
				tool.Config.CreateEntryAction = &toolsproto.ActionLink{
					ToolId: createToolId,
					Data:   []*toolsproto.DataMapping{},
				}
			}
		}
	}
}

// generateEmbeddedActionLinks will create links to embedded actions. These are generated for GET actions for models that
// have HasMany relationships provided that:
// - the related model has a list action that has the parent field as a filter
func (g *Generator) generateEmbeddedActionLinks() {
	for _, tool := range g.Tools {
		if !tool.Action.IsGet() {
			continue
		}

		for _, f := range tool.Model.Fields {
			// skip if the field is not a HasMany relationship
			if !f.IsHasMany() {
				continue
			}

			// find the list tools for the related model
			listTools := g.findListTools(f.Type.ModelName.Value)
			for _, toolId := range listTools {
				// check if there is an input for the foreign key
				if input := g.Tools[toolId].getInput("$.where." + f.InverseFieldName.Value + ".id.equals"); input != nil {
					// embed the tool
					tool.Config.EmbeddedActions = append(tool.Config.EmbeddedActions, &toolsproto.ActionLink{
						ToolId: toolId,
						Title:  &toolsproto.StringTemplate{Template: f.Name}, // e.g. `orderItems` on a getOrder action
						Data: []*toolsproto.DataMapping{
							{
								Key:  input.FieldLocation.Path,
								Path: &toolsproto.JsonPath{Path: tool.getIDResponseFieldPath()},
							},
						},
					})

					break
				}
			}
		}
	}
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

		fields, err := g.makeInputsForMessage(tool.Action.Type, msg, "")
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
			// appent wrapper response for the results
			tool.Config.Response = append(tool.Config.Response, &toolsproto.ResponseFieldConfig{
				FieldLocation: &toolsproto.JsonPath{Path: "$.results"},
				FieldType:     proto.Type_TYPE_OBJECT,
				Repeated:      true,
				DisplayName:   "Results",
				Visible:       true,
			})
		}
		fields, err := g.makeResponsesForModel(tool.Model, pathPrefix, tool.Action.GetResponseEmbeds(), tool.SortableFields)
		if err != nil {
			return err
		}
		tool.Config.Response = append(tool.Config.Response, fields...)
	}

	return nil
}

func (g *Generator) makeInputsForMessage(actionType proto.ActionType, msg *proto.Message, pathPrefix string) ([]*toolsproto.RequestFieldConfig, error) {
	fields := []*toolsproto.RequestFieldConfig{}

	for i, f := range msg.GetFields() {
		if f.IsMessage() {
			submsg := g.Schema.FindMessage(f.Type.MessageName.Value)
			if submsg == nil {
				return nil, ErrInvalidSchema
			}

			fields = append(fields, &toolsproto.RequestFieldConfig{
				FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.Name},
				FieldType:     f.Type.Type,
				Repeated:      f.Type.Repeated,
				DisplayName:   casing.ToSentenceCase(f.Name),
				DisplayOrder:  int32(i),
				Visible:       true,
			})

			prefix := pathPrefix + "." + f.Name
			if f.Type.Repeated {
				prefix = prefix + "[*]"
			}

			subFields, err := g.makeInputsForMessage(actionType, submsg, prefix)
			if err != nil {
				return nil, err
			}
			fields = append(fields, subFields...)

			continue
		}

		config := &toolsproto.RequestFieldConfig{
			FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.Name},
			FieldType:     f.Type.Type,
			Repeated:      f.Type.Repeated,
			DisplayName:   casing.ToSentenceCase(f.Name),
			DisplayOrder:  int32(i),
			Visible:       true,
		}

		if f.Type.ModelName != nil && f.Type.FieldName != nil && proto.FindField(g.Schema.Models, f.Type.ModelName.Value, f.Type.FieldName.Value).Unique {
			// generate action link placeholders
			if lookupToolsIDs := g.findListTools(f.Type.ModelName.Value); len(lookupToolsIDs) > 0 {
				config.LookupAction = &toolsproto.ActionLink{
					ToolId: lookupToolsIDs[0],
				}
			}

			// create the GetEntry tool link to retrieve the entry for this related model. At this point, not all tools'
			// inputs and responses have been generated ; this is a placeholder that will have it's data populated later
			// in the generation process
			if entryToolID := g.findGetTool(f.Type.ModelName.Value); entryToolID != "" {
				// We do not add a GetEntryAction for the 'id' (or any unique lookup) input on a 'get', 'create' or 'update' action of a model, however do we add it for related models
				if !((actionType == proto.ActionType_ACTION_TYPE_GET || actionType == proto.ActionType_ACTION_TYPE_CREATE || actionType == proto.ActionType_ACTION_TYPE_UPDATE) && len(f.Target) == 1) {
					config.GetEntryAction = &toolsproto.ActionLink{
						ToolId: entryToolID,
					}
				}
			}
		}

		fields = append(fields, config)
	}

	return fields, nil
}

func (g *Generator) makeResponsesForMessage(msg *proto.Message, pathPrefix string, sortableFields []string) ([]*toolsproto.ResponseFieldConfig, error) {
	fields := []*toolsproto.ResponseFieldConfig{}
	order := 0
	for _, f := range msg.GetFields() {
		if f.IsMessage() {
			submsg := g.Schema.FindMessage(f.Type.MessageName.Value)
			if submsg == nil {
				return nil, ErrInvalidSchema
			}

			fields = append(fields, &toolsproto.ResponseFieldConfig{
				FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.Name},
				FieldType:     f.Type.Type,
				Repeated:      f.Type.Repeated,
				DisplayName:   casing.ToSentenceCase(f.Name),
				DisplayOrder:  computeFieldOrder(&order, len(msg.GetFields()), f.Name),
				Visible:       true,
			})

			prefix := pathPrefix + "." + f.Name
			if f.Type.Repeated {
				prefix = prefix + "[*]"
			}

			subFields, err := g.makeResponsesForMessage(submsg, prefix, []string{})
			if err != nil {
				return nil, err
			}
			fields = append(fields, subFields...)

			continue
		} else if f.Type.Type == proto.Type_TYPE_MODEL {
			model := g.Schema.FindModel(f.Type.ModelName.Value)
			if model == nil {
				return nil, ErrInvalidSchema
			}

			fields = append(fields, &toolsproto.ResponseFieldConfig{
				FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.Name},
				FieldType:     f.Type.Type,
				Repeated:      f.Type.Repeated,
				DisplayName:   casing.ToSentenceCase(f.Name),
				DisplayOrder:  computeFieldOrder(&order, len(msg.GetFields()), f.Name),
				Visible:       true,
			})

			prefix := pathPrefix + "." + f.Name
			if f.Type.Repeated {
				prefix = prefix + "[*]"
			}

			subFields, err := g.makeResponsesForModel(model, prefix, []string{}, []string{})
			if err != nil {
				return nil, err
			}
			fields = append(fields, subFields...)

			continue
		}

		config := &toolsproto.ResponseFieldConfig{
			FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.Name},
			FieldType:     f.Type.Type,
			Repeated:      f.Type.Repeated,
			DisplayName:   casing.ToSentenceCase(f.Name),
			Visible:       true,
			DisplayOrder:  computeFieldOrder(&order, len(msg.GetFields()), f.Name),
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
func (g *Generator) makeResponsesForModel(model *proto.Model, pathPrefix string, embeddings []string, sortableFields []string) ([]*toolsproto.ResponseFieldConfig, error) {
	fields := []*toolsproto.ResponseFieldConfig{}
	order := 0

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

		config := &toolsproto.ResponseFieldConfig{
			FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.Name},
			FieldType:     f.Type.Type,
			Repeated:      f.Type.Repeated,
			DisplayName: func() string {
				// if the field is a model (relationship), the display name of the field should be the
				// name of the related field without the ID suffix; e.g. "Category" instead of "Category id"
				if f.IsForeignKey() {
					return casing.ToSentenceCase(strings.TrimSuffix(f.Name, "Id"))
				}
				return casing.ToSentenceCase(f.Name)
			}(),
			Visible:      true,
			DisplayOrder: computeFieldOrder(&order, len(model.GetFields()), f.Name),
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
				config.Link = &toolsproto.ActionLink{
					ToolId: getToolID,
					Data: []*toolsproto.DataMapping{
						{
							Key:  g.Tools[getToolID].getIDInputFieldPath(),
							Path: config.FieldLocation,
						},
					},
				}
			}
		}

		fields = append(fields, config)
	}

	return fields, nil
}

func computeFieldOrder(currentOrder *int, fieldCount int, fieldName string) int32 {
	switch fieldName {
	case "id":
		return int32(fieldCount - 2)
	case "createdAt":
		return int32(fieldCount - 1)
	case "updatedAt":
		return int32(fieldCount)
	}
	val := *currentOrder
	*currentOrder++
	return int32(val)
}

// findListTools will search for list tools for the given model
func (g *Generator) findListTools(modelName string) []string {
	ids := []string{}
	for id, tool := range g.Tools {
		if tool.Model.Name == modelName && tool.Action.IsList() {
			ids = append(ids, id)
		}
	}

	sort.Strings(ids)

	return ids
}

// findGetTool will search for a get tool for the given model. It will prioritise a get(id) action.
func (g *Generator) findGetTool(modelName string) string {
	for id, tool := range g.Tools {
		if tool.Model.Name == modelName && tool.Action.IsGet() && tool.hasOnlyIDInput() {
			return id
		}
	}

	for id, tool := range g.Tools {
		if tool.Model.Name == modelName && tool.Action.IsGet() {
			return id
		}
	}

	return ""
}

// findCreateTool will search for a get tool for the given model
func (g *Generator) findCreateTool(modelName string) string {
	for id, tool := range g.Tools {
		if tool.Model.Name == modelName && tool.Action.IsCreate() {
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

// findAllByIDTools searches for the tools that operate on the given model and take in an ID as an input; Returns a map of
// tool IDs and the path of the input field; e.g. getPost: $.id. Results will omit the given tool id (ignoreID).
//
// GET READ DELETE WRITE etc tools are included if they take in only on input (the ID)
// UPDATE tools are included if they take in a where.id input alongside other inputs
func (g *Generator) findAllByIDTools(modelName string, ignoreID string) map[string]string {
	toolIds := map[string]string{}
	for id, tool := range g.Tools {
		if id == ignoreID {
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
func (g *Generator) makeCapabilities(action *proto.Action) *toolsproto.Capabilities {
	c := &toolsproto.Capabilities{
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
//
// If no text fields found, we revert to the sentence-cased action name (also removing the list/get/read prefixes
// (e.g. Invoices instead of List invoices)
func (g *Generator) makeTitle(action *proto.Action, model *proto.Model) *toolsproto.StringTemplate {
	if action.IsGet() || action.Type == proto.ActionType_ACTION_TYPE_READ {
		fields := model.GetFields()
		if len(fields) > 0 && fields[0].Type.Type == proto.Type_TYPE_STRING {
			return &toolsproto.StringTemplate{
				Template: "{{$." + fields[0].GetName() + "}}",
			}
		}
	}

	actionName := action.Name

	switch action.Type {
	case proto.ActionType_ACTION_TYPE_GET:
		actionName = strings.TrimPrefix(action.Name, "get")
	case proto.ActionType_ACTION_TYPE_LIST:
		actionName = strings.TrimPrefix(action.Name, "list")
	case proto.ActionType_ACTION_TYPE_READ:
		actionName = strings.TrimPrefix(action.Name, "read")
	}

	return &toolsproto.StringTemplate{
		Template: casing.ToSentenceCase(actionName),
	}
}

// getPageInfoResponses will return the responses for pageInfo (by default available on all autogenerated LIST actions)
func getPageInfoResponses() []*toolsproto.ResponseFieldConfig {
	return []*toolsproto.ResponseFieldConfig{
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo"},
			FieldType:     proto.Type_TYPE_OBJECT,
			DisplayName:   "PageInfo",
			Visible:       false,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.count"},
			FieldType:     proto.Type_TYPE_INT,
			DisplayName:   "Count",
			Visible:       false,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.totalCount"},
			FieldType:     proto.Type_TYPE_INT,
			DisplayName:   "Total count",
			Visible:       false,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.hasNextPage"},
			FieldType:     proto.Type_TYPE_BOOL,
			DisplayName:   "Has next page",
			Visible:       false,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.startCursor"},
			FieldType:     proto.Type_TYPE_STRING,
			DisplayName:   "Start cursor",
			Visible:       false,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.endCursor"},
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
		if !(input.FieldType == proto.Type_TYPE_ID && input.FieldLocation.String() == "$.id") {
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

// getInput finds and returns an inpub by it's path; returns nil if not found
func (t *Tool) getInput(path string) *toolsproto.RequestFieldConfig {
	for _, input := range t.Config.Inputs {
		if input.FieldLocation.Path == path {
			return input
		}
	}

	return nil
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
