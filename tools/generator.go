package tools

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema/parser"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

type Tool struct {
	ID string

	// For tools powered by actions
	ActionConfig   *toolsproto.ActionConfig
	Model          *proto.Model
	Action         *proto.Action
	SortableFields []string

	// For tools powered by flows
	FlowConfig *toolsproto.FlowConfig
	Flow       *proto.Flow
}

func (t *Tool) IsActionBased() bool {
	return t.ActionConfig != nil && t.Action != nil
}

func (t *Tool) IsFlowBased() bool {
	return t.FlowConfig != nil && t.Flow != nil
}

func (t *Tool) asPB() *toolsproto.Tool {
	if t.IsActionBased() {
		return &toolsproto.Tool{
			Id:           t.ID,
			Type:         toolsproto.Tool_ACTION,
			ActionConfig: t.ActionConfig,
		}
	}
	if t.IsFlowBased() {
		return &toolsproto.Tool{
			Id:         t.ID,
			Type:       toolsproto.Tool_FLOW,
			FlowConfig: t.FlowConfig,
		}
	}

	return nil
}

type FieldType string

const (
	FieldTypeModel FieldType = "MODEL"
	FieldTypeEnum  FieldType = "ENUM"
)

// Field represents a model field or an enum. These fields have formatting configuration that is generated and
// can be configured by users.
type Field struct {
	Type      FieldType
	EnumName  string
	ModelName string
	FieldName string
}

func (f *Field) Path() string {
	switch f.Type {
	case FieldTypeEnum:
		return f.EnumName
	case FieldTypeModel:
		return f.ModelName + "." + f.FieldName
	default:
		return ""
	}
}

func (f *Field) asPB() *toolsproto.Field {
	if f == nil {
		return nil
	}

	return &toolsproto.Field{
		Type: func() toolsproto.Field_Type {
			switch f.Type {
			case FieldTypeEnum:
				return toolsproto.Field_ENUM
			default:
				return toolsproto.Field_MODEL
			}
		}(),
		EnumName:  stringPointer(f.EnumName),
		ModelName: stringPointer(f.ModelName),
		FieldName: stringPointer(f.FieldName),
	}
}

type Fields map[string]*Field

type Generator struct {
	Schema     *proto.Schema
	KeelConfig *config.ProjectConfig
	Tools      map[string]*Tool
	// Fields represents the complete set of model fields and enums that are present in the given schema
	Fields Fields
}

func (g *Generator) actionTools() map[string]*Tool {
	actionTools := map[string]*Tool{}
	for id, tool := range g.Tools {
		if tool.IsActionBased() {
			actionTools[id] = tool
		}
	}
	return actionTools
}

func (g *Generator) flowTools() map[string]*Tool {
	flowTools := map[string]*Tool{}
	for id, tool := range g.Tools {
		if tool.IsFlowBased() {
			flowTools[id] = tool
		}
	}
	return flowTools
}

const fieldNameID = "id"

var ErrInvalidSchema = errors.New("invalid schema")

// NewGenerator creates a new tool config generator for the given schema.
func NewGenerator(schema *proto.Schema, keelConfig *config.ProjectConfig) (*Generator, error) {
	return &Generator{
		Schema:     schema,
		KeelConfig: keelConfig,
	}, nil
}

// GetTools returns all the tools that have been generated in alphabetical order. These include both action and flow tools.
func (g *Generator) GetTools() []*toolsproto.Tool {
	tools := []*toolsproto.Tool{}
	ids := []string{}
	for id := range g.Tools {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		tools = append(tools, g.Tools[id].asPB())
	}

	return tools
}

// GetFields returns all the fields that have been generated in alphabetical order.
func (g *Generator) GetFields() []*toolsproto.Field {
	fields := []*toolsproto.Field{}
	ids := []string{}
	for id := range g.Fields {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		fields = append(fields, g.Fields[id].asPB())
	}

	return fields
}

// Generate will generate all the tools for this generator's schema.
func (g *Generator) Generate(ctx context.Context) error {
	if g.Schema == nil {
		return ErrInvalidSchema
	}

	// reset any previous tools & fields
	g.Tools = map[string]*Tool{}

	// first pass at generating tools;
	g.scaffoldTools()

	// then decorate the tools with all the relevant options
	if err := g.decorateTools(); err != nil {
		return fmt.Errorf("decorating tools: %w", err)
	}

	return nil
}

// Generate will generate all the fields for this generator's schema.
// It will take the current schema and generate the map of all model fields & enums that can be formatted with
// schema-based config.
func (g *Generator) GenerateFields(ctx context.Context) error {
	if g.Schema == nil {
		return ErrInvalidSchema
	}

	// reset any previous fields
	g.Fields = Fields{}

	// generate all the fields
	for _, model := range g.Schema.GetModels() {
		for _, field := range model.GetFields() {
			f := Field{
				Type:      FieldTypeModel,
				ModelName: field.GetEntityName(),
				FieldName: field.GetName(),
				EnumName:  field.GetType().GetEnumName().GetValue(),
			}
			g.Fields[f.Path()] = &f
		}
	}

	for _, enum := range g.Schema.GetEnums() {
		f := Field{
			Type:     FieldTypeEnum,
			EnumName: enum.GetName(),
		}
		g.Fields[f.Path()] = &f
	}

	return nil
}

// scaffoldTools will generate all the basic tools. These will be incomplete configurations, with fields and
// relations between them not yet filled in
//
// For each model's actions, we will scaffold the `ActionConfig`s. These will not yet contain all request fields,
// response fields and any related/embedded tools, as these need to reference each other, so we first scaffold them and
// the completed generation is done later on
//
// For each flow we will scaffold the `FlowConfig`s.
func (g *Generator) scaffoldTools() {
	var api *proto.Api
	if g.KeelConfig != nil && g.KeelConfig.Console.Api != nil {
		if api = proto.FindApi(g.Schema, *g.KeelConfig.Console.Api); api == nil {
			return
		}
	} else {
		if api = proto.FindApi(g.Schema, parser.DefaultApi); api == nil {
			return
		}
	}

	for _, model := range g.Schema.GetModels() {
		for _, action := range model.GetActions() {
			if !slices.Contains(proto.GetActionNamesForApi(g.Schema, api), action.GetName()) {
				continue
			}

			t := Tool{
				ID: casing.ToKebab(action.GetName()),
				ActionConfig: &toolsproto.ActionConfig{
					Id:             casing.ToKebab(action.GetName()),
					ApiNames:       g.Schema.FindApiNames(model.GetName(), action.GetName()),
					Name:           casing.ToSentenceCase(action.GetName()),
					ActionName:     action.GetName(),
					ModelName:      model.GetName(),
					ActionType:     action.GetType(),
					Implementation: action.GetImplementation(),
					EntitySingle:   strings.ToLower(casing.ToSentenceCase(model.GetName())),
					EntityPlural:   casing.ToPlural(strings.ToLower(casing.ToSentenceCase(model.GetName()))),
					Capabilities:   g.makeCapabilities(action),
					Title:          g.makeTitle(action, model),
					FilterConfig:   g.makeFilterConfig(),
				},
				Model:  model,
				Action: action,
			}

			defaultPageSize := int32(50)
			// List actions have pagination
			if action.IsList() {
				t.ActionConfig.Pagination = &toolsproto.CursorPaginationConfig{
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
						DefaultValue:  &defaultPageSize,
					},
					NextPage:   &toolsproto.JsonPath{Path: "$.pageInfo.hasNextPage"},
					TotalCount: &toolsproto.JsonPath{Path: "$.pageInfo.totalCount"},
				}
			}

			// get actions have a display layout of RecordView
			if action.IsGet() {
				t.ActionConfig.DisplayLayout = &toolsproto.DisplayLayoutConfig{
					Type:         toolsproto.DisplayLayoutConfig_RECORD,
					RecordConfig: &toolsproto.RecordViewConfig{},
				}
			}

			g.Tools[t.ID] = &t
		}
	}

	for _, flow := range g.Schema.GetFlows() {
		t := Tool{
			ID: casing.ToKebab(flow.GetName()),
			FlowConfig: &toolsproto.FlowConfig{
				Name:     casing.ToSentenceCase(flow.GetName()),
				FlowName: flow.GetName(),
			},
			Flow: flow,
		}

		if model := inferFlowRelatedModel(g.Schema, flow); model != nil {
			t.Model = model
			t.FlowConfig.ModelName = model.GetName()
		}
		g.Tools[t.ID] = &t
	}
}

func (g *Generator) decorateTools() error {
	if err := g.generateInputs(); err != nil {
		return fmt.Errorf("generating inputs for action based tools: %w", err)
	}

	if err := g.generateFlowInputs(); err != nil {
		return fmt.Errorf("generating inputs for flow based tools: %w", err)
	}

	if err := g.generateResponses(); err != nil {
		return fmt.Errorf("generating responses: %w", err)
	}

	g.generateRelatedActionsLinks()
	g.generateEntryActivityActionsLinks()
	g.generateGetEntryActionLinks()
	g.generateEmbeddedTools()
	g.generateCreateEntryActionLinks()

	// decorate further...
	for _, tool := range g.actionTools() {
		// for all inputs that are IDs that have a get_entry_action link (e.g. used to lookup a related model field),
		// find the get(id) tool and decorate the data mapping now that we have all inputs and responses generated
		for _, input := range tool.ActionConfig.GetInputs() {
			if input.GetGetEntryAction() != nil && input.GetGetEntryAction().GetToolId() != "" {
				if entryToolID := g.findGetByIDTool(input.GetGetEntryAction().GetToolId()); entryToolID != "" {
					input.GetEntryAction.ToolId = entryToolID
					input.GetEntryAction.Data = []*toolsproto.DataMapping{
						{
							Key:  g.Tools[entryToolID].getIDInputFieldPath(),
							Path: input.GetFieldLocation(),
						},
					}
				} else {
					// if not get(id) is found, then remove the GetEntryAction placeholder
					input.GetEntryAction = nil
				}
			}
		}

		// for all responses that have a link for to-many fields,
		// decorate the data mapping now that we have all inputs and responses generated
		for _, response := range tool.ActionConfig.GetResponse() {
			if !tool.Action.IsArbitraryFunction() && response.GetLink() != nil && response.GetLink().GetToolId() != "" && response.GetLink().GetData()[0].GetPath() == nil {
				response.Link.Data[0].Path = &toolsproto.JsonPath{
					Path: tool.getIDResponseFieldPath(),
				}
			}
		}
	}

	return nil
}

// generateRelatedActionsLinks will traverse the tools and generate the RelatedActions links:
//   - For LIST actions = other list actions for the same model
//   - For DELETE actions = all list actions for the same model
func (g *Generator) generateRelatedActionsLinks() {
	for id, tool := range g.actionTools() {
		displayOrder := 0
		if !(tool.Action.IsList() || tool.Action.IsDelete()) {
			continue
		}

		// we search for more than one list tool as the results will include the one we're on
		if relatedTools := g.findListTools(tool.Model.GetName()); len(relatedTools) > 1 {
			for _, relatedID := range relatedTools {
				if id != relatedID {
					displayOrder++
					tool.ActionConfig.RelatedActions = append(tool.ActionConfig.RelatedActions, &toolsproto.ToolLink{
						ToolId:       relatedID,
						DisplayOrder: int32(displayOrder),
					})
				}
			}
		}
	}
}

// generateEntryActivityActionsLinks will traverse the tools and generate the EntryActivityActions links:
//   - For LIST/GET actions that have a model ID response = other actions on the same model that take an id as an input
func (g *Generator) generateEntryActivityActionsLinks() {
	for id, tool := range g.actionTools() {
		displayOrder := 0
		// get the path of the id response field for this tool
		idResponseFieldPath := tool.getIDResponseFieldPath()
		// skip if we don't have an id response field or the tool is not List or Get
		if idResponseFieldPath == "" || (!tool.Action.IsList() && !tool.Action.IsGet()) {
			continue
		}

		// entry activity actions for GET and LIST that have an id response
		inputPaths := g.findAllByIDTools(tool.Model.GetName(), id)

		// now we sort the tools by name But with deletes at the end
		ids := []string{}
		for toolID := range inputPaths {
			ids = append(ids, toolID)
		}
		slices.SortFunc(ids, func(a, b string) int {
			if strings.HasPrefix(a, "delete") {
				if strings.HasPrefix(b, "delete") {
					return 0
				}

				return 1
			}

			if strings.HasPrefix(b, "delete") {
				return -1
			}

			return strings.Compare(a, b)
		})

		for _, toolID := range ids {
			displayOrder++

			asDialog := false

			targetTool := g.Tools[toolID]
			if targetTool.Action.IsWriteAction() {
				asDialog = true
			}

			tool.ActionConfig.EntryActivityActions = append(tool.ActionConfig.EntryActivityActions, &toolsproto.ToolLink{
				ToolId: toolID,
				Data: []*toolsproto.DataMapping{
					{
						Key:  inputPaths[toolID],
						Path: &toolsproto.JsonPath{Path: idResponseFieldPath},
					},
				},
				DisplayOrder: int32(displayOrder),
				AsDialog:     &asDialog,
			})
		}

		// now we look for flows related to this tool's model
		for _, ft := range g.flowTools() {
			asDialog := true
			if ft.Model.GetName() == tool.Model.GetName() {
				displayOrder++
				tool.ActionConfig.EntryActivityActions = append(tool.ActionConfig.EntryActivityActions, &toolsproto.ToolLink{
					ToolId:       ft.ID,
					DisplayOrder: int32(displayOrder),
					AsDialog:     &asDialog,
				})
			}
		}
	}
}

// generateGetEntryActionLinks will traverse the tools and generate the GetEntryAction links:
//   - For LIST/UPDATE/CREATE = a GET action used to retrieve the model by id
func (g *Generator) generateGetEntryActionLinks() {
	for _, tool := range g.actionTools() {
		// get the path of the id response field for this tool
		idResponseFieldPath := tool.getIDResponseFieldPath()
		if idResponseFieldPath == "" {
			continue
		}
		// get entry action for tools that operate on a model instance/s (create/update/list).
		if tool.Action.IsList() || tool.Action.IsUpdate() || tool.Action.GetType() == proto.ActionType_ACTION_TYPE_CREATE {
			if getToolID := g.findGetByIDTool(tool.Model.GetName()); getToolID != "" {
				tool.ActionConfig.GetEntryAction = &toolsproto.ToolLink{
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
	for _, tool := range g.actionTools() {
		if tool.Action.IsList() || tool.Action.IsGet() {
			if createToolId := g.findCreateTool(tool.Model.GetName()); createToolId != "" {
				//TODO: improvement: add datamapping from list actions to the create action if there are any filtered fields
				tool.ActionConfig.CreateEntryAction = &toolsproto.ToolLink{
					ToolId: createToolId,
					Data:   []*toolsproto.DataMapping{},
				}
			}
		}
	}
}

// generateEmbeddedTools will create links to embedded actions. These are generated for GET actions for models that
// have HasMany relationships provided that:
// - the related model has a list action that has the parent field as a filter.
func (g *Generator) generateEmbeddedTools() {
	for _, tool := range g.actionTools() {
		if !tool.Action.IsGet() {
			continue
		}
		displayOrder := 0

		for _, f := range tool.Model.GetFields() {
			// skip if the field is not a HasMany relationship
			if !f.IsHasMany() {
				continue
			}

			// find the list tools for the related model
			listTools := g.findListTools(f.GetType().GetEntityName().GetValue())
			for _, toolId := range listTools {
				// check if there is an input for the foreign key
				if input := g.Tools[toolId].getInput("$.where." + f.GetInverseFieldName().GetValue() + ".id.equals"); input != nil {
					displayOrder++
					// embed the tool as a tool group
					tool.ActionConfig.EmbeddedTools = append(tool.ActionConfig.EmbeddedTools, &toolsproto.ToolGroup{
						Id:           f.GetName(),
						Title:        &toolsproto.StringTemplate{Template: casing.ToSentenceCase(f.GetName())}, // e.g. `Order items` on a getOrder action
						DisplayOrder: int32(displayOrder),
						Tools: []*toolsproto.ToolGroup_GroupActionLink{
							{
								ActionLink: &toolsproto.ToolLink{
									ToolId: toolId,
									Title:  &toolsproto.StringTemplate{Template: f.GetName()}, // e.g. `orderItems` on a getOrder action
									Data: []*toolsproto.DataMapping{
										{
											Key:  input.GetFieldLocation().GetPath(),
											Path: &toolsproto.JsonPath{Path: tool.getIDResponseFieldPath()},
										},
									},
								},
								ResponseOverrides: []*toolsproto.ResponseOverrides{{
									FieldLocation: &toolsproto.JsonPath{Path: "$.results[*]." + f.GetInverseFieldName().GetValue() + "Id"},
									Visible:       false,
								}},
							},
						},
						Visible: true,
					})

					break
				}
			}
		}
	}
}

// generateInputs will make the inputs for all action based tools.
func (g *Generator) generateInputs() error {
	for _, tool := range g.actionTools() {
		// if the action does not have a input message, it means we don't have any inputs for this tool
		if tool.Action.GetInputMessageName() == "" {
			continue
		}

		// get the input message
		msg := g.Schema.FindMessage(tool.Action.GetInputMessageName())
		if msg == nil {
			return ErrInvalidSchema
		}

		fields, err := g.makeInputsForMessage(tool.Action.GetType(), msg, "", nil)
		if err != nil {
			return err
		}
		tool.ActionConfig.Inputs = fields

		// If there are any OrderBy fields, then we find the sortable field names and store them against the tool, to be
		// used later on when generating the response
		if orderBy := msg.GetOrderByField(); orderBy != nil {
			sortableFields := []string{}
			for _, unionMsgName := range orderBy.GetType().GetUnionNames() {
				unionMsg := g.Schema.FindMessage(unionMsgName.GetValue())
				if unionMsg == nil {
					return ErrInvalidSchema
				}
				for _, f := range unionMsg.GetFields() {
					if f.GetType().GetType() == proto.Type_TYPE_SORT_DIRECTION {
						sortableFields = append(sortableFields, f.GetName())
					}
				}
			}

			tool.SortableFields = sortableFields
		}
	}

	return nil
}

// generateFlowInputs will make the inputs for all flow based tools.
func (g *Generator) generateFlowInputs() error {
	for _, tool := range g.flowTools() {
		// if the flow does not have a input message, it means we don't have any inputs for this tool
		if tool.Flow.GetInputMessageName() == "" {
			continue
		}

		// get the input message
		msg := g.Schema.FindMessage(tool.Flow.GetInputMessageName())
		if msg == nil {
			return ErrInvalidSchema
		}

		fields := []*toolsproto.FlowInputConfig{}

		for i, f := range msg.GetFields() {
			if f.IsMessage() {
				// nested messages aren't supported for flows
				continue
			}

			config := &toolsproto.FlowInputConfig{
				FieldLocation: &toolsproto.JsonPath{Path: `$.` + f.GetName()},
				FieldType:     f.GetType().GetType(),
				Repeated:      f.GetType().GetRepeated(),
				DisplayName:   casing.ToSentenceCase(f.GetName()),
				DisplayOrder:  int32(i),
			}

			if f.GetType().GetEntityName() != nil {
				config.ModelName = &f.Type.EntityName.Value
			}
			if f.GetType().GetFieldName() != nil {
				config.FieldName = &f.Type.FieldName.Value
			}
			if f.GetType().GetEnumName() != nil {
				config.EnumName = &f.Type.EnumName.Value
			}

			fields = append(fields, config)
		}

		tool.FlowConfig.Inputs = fields
	}

	return nil
}

// generateResponses will make the responses for all action based tools.
func (g *Generator) generateResponses() error {
	for _, tool := range g.actionTools() {
		// skip tools that are flow based
		if tool.IsFlowBased() {
			continue
		}

		// if the action has a response message, let's generate it
		if tool.Action.GetResponseMessageName() != "" {
			// get the message
			msg := g.Schema.FindMessage(tool.Action.GetResponseMessageName())
			if msg == nil {
				return ErrInvalidSchema
			}

			fields, err := g.makeResponsesForMessage(msg, "", tool.SortableFields)
			if err != nil {
				return err
			}
			tool.ActionConfig.Response = fields

			continue
		}

		// delete actions do not have a response
		if tool.Action.IsDelete() {
			continue
		}

		// we don't have a response message, therefore the response will be the model...
		pathPrefix := ""
		// if the action is a list action, we also need to include the pageInfo responses, resultInfo responses and prefix the results
		if tool.Action.IsList() {
			pathPrefix = ".results[*]"
			tool.ActionConfig.Response = append(tool.ActionConfig.Response, getPageInfoResponses()...)

			if len(tool.Action.GetFacets()) > 0 {
				resultInfo, err := getResultInfoResponses(g.Schema, tool.Action)
				if err != nil {
					return err
				}
				tool.ActionConfig.Response = append(tool.ActionConfig.Response, resultInfo...)
			}
		}
		fields, err := g.makeResponsesForModel(tool.Model, pathPrefix, tool.Action.GetResponseEmbeds(), tool.SortableFields)
		if err != nil {
			return err
		}
		tool.ActionConfig.Response = append(tool.ActionConfig.Response, fields...)
	}

	return nil
}

func (g *Generator) makeInputsForMessage(
	actionType proto.ActionType,
	msg *proto.Message,
	pathPrefix string,
	scope *toolsproto.RequestFieldConfig_ScopeType,
) ([]*toolsproto.RequestFieldConfig, error) {
	fields := []*toolsproto.RequestFieldConfig{}

	for i, f := range msg.GetFields() {
		var fScope toolsproto.RequestFieldConfig_ScopeType
		if scope == nil {
			fScope = inferInputType(actionType, f.GetName())
		} else {
			fScope = *scope
		}
		if f.IsMessage() {
			submsg := g.Schema.FindMessage(f.GetType().GetMessageName().GetValue())
			if submsg == nil {
				return nil, ErrInvalidSchema
			}

			fields = append(fields, &toolsproto.RequestFieldConfig{
				Scope:         fScope,
				FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.GetName()},
				FieldType:     f.GetType().GetType(),
				Repeated:      f.GetType().GetRepeated(),
				DisplayName:   casing.ToSentenceCase(f.GetName()),
				DisplayOrder:  int32(i),
				Visible:       true,
			})

			prefix := pathPrefix + "." + f.GetName()
			if f.GetType().GetRepeated() {
				prefix = prefix + "[*]"
			}

			subFields, err := g.makeInputsForMessage(actionType, submsg, prefix, &fScope)
			if err != nil {
				return nil, err
			}
			fields = append(fields, subFields...)

			continue
		}

		config := &toolsproto.RequestFieldConfig{
			Scope:         fScope,
			FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.GetName()},
			FieldType:     f.GetType().GetType(),
			Repeated:      f.GetType().GetRepeated(),
			DisplayName:   casing.ToSentenceCase(f.GetName()),
			DisplayOrder:  int32(i),
			Visible:       true,
		}

		if f.GetType().GetEntityName() != nil {
			config.ModelName = &f.Type.EntityName.Value
		}
		if f.GetType().GetFieldName() != nil {
			config.FieldName = &f.Type.FieldName.Value
		}
		if f.GetType().GetEnumName() != nil {
			config.EnumName = &f.Type.EnumName.Value
		}

		if f.GetType().GetEntityName() != nil && f.GetType().GetFieldName() != nil && g.Schema.FindEntity(f.GetType().GetEntityName().GetValue()).FindField(f.GetType().GetFieldName().GetValue()).GetUnique() {
			// generate lookup action only for ID inputs
			if f.GetType().GetFieldName().GetValue() == fieldNameID {
				if lookupToolsIDs := g.findListTools(f.GetType().GetEntityName().GetValue()); len(lookupToolsIDs) > 0 {
					config.LookupAction = &toolsproto.ToolLink{
						ToolId: lookupToolsIDs[0],
					}
				}
			}

			// create the GetEntry tool link to retrieve the entry for this related model. At this point, not all tools'
			// inputs and responses have been generated ; this is a placeholder that will have it's data populated later
			// in the generation process
			// We do not add a GetEntryAction for the 'id' (or any unique lookup) input on a 'get', 'create' or 'update' action of a model, however do we add it for related models
			if !((actionType == proto.ActionType_ACTION_TYPE_GET || actionType == proto.ActionType_ACTION_TYPE_CREATE || actionType == proto.ActionType_ACTION_TYPE_UPDATE) && len(f.GetTarget()) == 1) {
				config.GetEntryAction = &toolsproto.ToolLink{
					ToolId: f.GetType().GetEntityName().GetValue(), // TODO: this is a bit of a hack placeholder because we do not know the underlying model which the field is pointing to during post-processing
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
			submsg := g.Schema.FindMessage(f.GetType().GetMessageName().GetValue())
			if submsg == nil {
				return nil, ErrInvalidSchema
			}

			fields = append(fields, &toolsproto.ResponseFieldConfig{
				Scope:         toolsproto.ResponseFieldConfig_DEFAULT,
				FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.GetName()},
				FieldType:     f.GetType().GetType(),
				Repeated:      f.GetType().GetRepeated(),
				DisplayName:   casing.ToSentenceCase(f.GetName()),
				DisplayOrder:  computeFieldOrder(&order, len(msg.GetFields()), f.GetName()),
				Visible:       true,
			})

			prefix := pathPrefix + "." + f.GetName()
			if f.GetType().GetRepeated() {
				prefix = prefix + "[*]"
			}

			subFields, err := g.makeResponsesForMessage(submsg, prefix, []string{})
			if err != nil {
				return nil, err
			}
			fields = append(fields, subFields...)

			continue
		} else if f.GetType().GetType() == proto.Type_TYPE_ENTITY {
			model := g.Schema.FindModel(f.GetType().GetEntityName().GetValue())
			if model == nil {
				return nil, ErrInvalidSchema
			}

			fields = append(fields, &toolsproto.ResponseFieldConfig{
				Scope:         toolsproto.ResponseFieldConfig_DEFAULT,
				FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.GetName()},
				FieldType:     f.GetType().GetType(),
				Repeated:      f.GetType().GetRepeated(),
				DisplayName:   casing.ToSentenceCase(f.GetName()),
				DisplayOrder:  computeFieldOrder(&order, len(msg.GetFields()), f.GetName()),
				Visible:       true,
			})

			prefix := pathPrefix + "." + f.GetName()
			if f.GetType().GetRepeated() {
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
			Scope:         toolsproto.ResponseFieldConfig_DEFAULT,
			FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.GetName()},
			FieldType:     f.GetType().GetType(),
			Repeated:      f.GetType().GetRepeated(),
			DisplayName:   casing.ToSentenceCase(f.GetName()),
			Visible:       true,
			DisplayOrder:  computeFieldOrder(&order, len(msg.GetFields()), f.GetName()),
			Sortable: func() bool {
				for _, fn := range sortableFields {
					if fn == f.GetName() {
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

// makeResponsesForModel will return an array of response fields for the given model.
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
				if frags[0] == f.GetName() {
					found = true
					// if we have to embed a child model for this field, we need to pass them through with the first segment removed
					if len(frags) > 1 {
						fieldEmbeddings = append(fieldEmbeddings, strings.Join(frags[1:], "."))
					}
				}
			}
			if found {
				prefix := pathPrefix + "." + f.GetName()
				if f.IsHasMany() {
					prefix = prefix + "[*]"
				}
				embeddedFields, err := g.makeResponsesForModel(g.Schema.FindModel(f.GetType().GetEntityName().GetValue()), prefix, fieldEmbeddings, []string{})
				if err != nil {
					return nil, err
				}
				fields = append(fields, embeddedFields...)
			}

			if !f.IsHasMany() {
				continue
			}
		}

		config := &toolsproto.ResponseFieldConfig{
			Scope:         toolsproto.ResponseFieldConfig_DEFAULT,
			FieldLocation: &toolsproto.JsonPath{Path: `$` + pathPrefix + "." + f.GetName()},
			FieldType:     f.GetType().GetType(),
			Repeated:      f.GetType().GetRepeated(),
			DisplayName: func() string {
				// if the field is a model (relationship), the display name of the field should be the
				// name of the related field without the ID suffix; e.g. "Category" instead of "Category id"
				if f.IsForeignKey() {
					return casing.ToSentenceCase(strings.TrimSuffix(f.GetName(), "Id"))
				}
				return casing.ToSentenceCase(f.GetName())
			}(),
			Visible:      true,
			DisplayOrder: computeFieldOrder(&order, len(model.GetFields()), f.GetName()),
			Sortable: func() bool {
				for _, fn := range sortableFields {
					if fn == f.GetName() {
						return true
					}
				}
				return false
			}(),
			ModelName: &f.EntityName,
			FieldName: &f.Name,
			EnumName: func() *string {
				if f.GetType().GetEnumName() != nil {
					return &f.Type.EnumName.Value
				}
				return nil
			}(),
		}

		if f.IsFile() {
			config.ImagePreview = true
		}

		// if this field is a model, we add a link to the action used to retrieve the related model. Note that inputs are
		// generated first, so we're safe to create a tool/action link now
		if f.IsForeignKey() && f.GetForeignKeyInfo().GetRelatedEntityField() == fieldNameID {
			if getToolID := g.findGetByIDTool(f.GetForeignKeyInfo().GetRelatedEntityName()); getToolID != "" {
				config.Link = &toolsproto.ToolLink{
					ToolId: getToolID,
					Data: []*toolsproto.DataMapping{
						{
							Key:  g.Tools[getToolID].getIDInputFieldPath(),
							Path: config.GetFieldLocation(),
						},
					},
				}
			}
		}

		if f.IsHasMany() {
			if getToolID, input := g.findListByForeignID(f.GetType().GetEntityName().GetValue(), f.GetInverseFieldName().GetValue()); getToolID != "" {
				config.Link = &toolsproto.ToolLink{
					ToolId: getToolID,
					Data: []*toolsproto.DataMapping{
						{
							Key: input.GetFieldLocation().GetPath(),
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

// findListTools will search for list tools for the given model.
func (g *Generator) findListTools(modelName string) []string {
	ids := []string{}
	for id, tool := range g.actionTools() {
		if tool.IsActionBased() && tool.Model.GetName() == modelName && tool.Action.IsList() {
			ids = append(ids, id)
		}
	}

	sort.Strings(ids)

	return ids
}

// findCreateTool will search for a get tool for the given model.
func (g *Generator) findCreateTool(modelName string) string {
	toolIds := []string{}
	for id, tool := range g.actionTools() {
		if tool.Model.GetName() == modelName && tool.Action.IsCreate() {
			toolIds = append(toolIds, id)
		}
	}

	if len(toolIds) > 0 {
		sort.Strings(toolIds)
		return toolIds[0]
	}

	return ""
}

// findGetByIDTool will search for a get tool for the given model that takes in an ID. It will prioritise a get(id) action without @embeds.
func (g *Generator) findGetByIDTool(modelName string) string {
	if id := g.findGetByIDWithoutEmbedsTool(modelName); id != "" {
		return id
	}

	for id, tool := range g.actionTools() {
		if tool.Model.GetName() == modelName && tool.Action.IsGet() && tool.hasOnlyIDInput() {
			return id
		}
	}

	return ""
}

// findGetByIDTool will search for a get tool for the given model that takes in an ID and has no @embeds defined.
func (g *Generator) findGetByIDWithoutEmbedsTool(modelName string) string {
	toolIds := []string{}
	for id, tool := range g.actionTools() {
		if tool.Model.GetName() == modelName && tool.Action.IsGet() && tool.hasOnlyIDInput() && len(tool.Action.GetResponseEmbeds()) == 0 {
			toolIds = append(toolIds, id)
		}
	}

	if len(toolIds) > 0 {
		sort.Strings(toolIds)
		return toolIds[0]
	}

	return ""
}

// findListByForeignID will search for a list tool for the given model which takes a specific foreign key as an input
// It will also return the request input field for that tool.
func (g *Generator) findListByForeignID(modelName string, inverseFieldName string) (string, *toolsproto.RequestFieldConfig) {
	for id, tool := range g.actionTools() {
		if input := tool.getInput("$.where." + inverseFieldName + ".id.equals"); tool.Model.GetName() == modelName && tool.Action.GetType() == proto.ActionType_ACTION_TYPE_LIST && input != nil {
			return id, input
		}
	}

	return "", nil
}

// findAllByIDTools searches for the tools that operate on the given model and take in an ID as an input; Returns a map of
// tool IDs and the path of the input field; e.g. getPost: $.id. Results will omit the given tool id (ignoreID).
//
// GET READ DELETE WRITE etc tools are included if they take in only on input (the ID)
// UPDATE tools are included if they take in a where.id input alongside other inputs.
func (g *Generator) findAllByIDTools(modelName string, ignoreID string) map[string]string {
	inputPaths := map[string]string{}

	for id, tool := range g.actionTools() {
		if id == ignoreID {
			continue
		}
		if tool.Model.GetName() != modelName {
			continue
		}

		// if we only have one input, an ID, add and continue
		if tool.hasOnlyIDInput() {
			inputPaths[id] = tool.getIDInputFieldPath()
			continue
		}

		// if we have a UPDATE that includes a where.ID
		if tool.Action.IsUpdate() {
			idInputPath := ""
			for _, input := range tool.ActionConfig.GetInputs() {
				if input.GetFieldType() == proto.Type_TYPE_ID && input.GetFieldLocation().GetPath() == "$.where.id" {
					idInputPath = input.GetFieldLocation().GetPath()
				}
			}
			if idInputPath != "" {
				inputPaths[id] = idInputPath
				continue
			}
		}
	}

	return inputPaths
}

// makeCapabilities generates the makeCapabilities/features available for a tool generated for the given action.
// Audit trail is enabled just for GET actions
// Comments are enabled just for GET actions.
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
// (e.g. Invoices instead of List invoices).
func (g *Generator) makeTitle(action *proto.Action, model *proto.Model) *toolsproto.StringTemplate {
	if action.IsGet() || action.GetType() == proto.ActionType_ACTION_TYPE_READ {
		fields := model.GetFields()
		if len(fields) > 0 && fields[0].GetType().GetType() == proto.Type_TYPE_STRING {
			return &toolsproto.StringTemplate{
				Template: "{{$." + fields[0].GetName() + "}}",
			}
		}
	}

	actionName := action.GetName()

	switch action.GetType() {
	case proto.ActionType_ACTION_TYPE_GET:
		actionName = strings.TrimPrefix(action.GetName(), "get")
	case proto.ActionType_ACTION_TYPE_LIST:
		actionName = strings.TrimPrefix(action.GetName(), "list")
	case proto.ActionType_ACTION_TYPE_READ:
		actionName = strings.TrimPrefix(action.GetName(), "read")
	}

	return &toolsproto.StringTemplate{
		Template: casing.ToSentenceCase(actionName),
	}
}

func (g *Generator) makeFilterConfig() *toolsproto.FilterConfig {
	return &toolsproto.FilterConfig{}
}

// getPageInfoResponses will return the responses for pageInfo (by default available on all autogenerated LIST actions).
func getPageInfoResponses() []*toolsproto.ResponseFieldConfig {
	return []*toolsproto.ResponseFieldConfig{
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo"},
			FieldType:     proto.Type_TYPE_OBJECT,
			DisplayName:   "PageInfo",
			Visible:       false,
			Scope:         toolsproto.ResponseFieldConfig_PAGINATION,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.count"},
			FieldType:     proto.Type_TYPE_INT,
			DisplayName:   "Count",
			Visible:       false,
			Scope:         toolsproto.ResponseFieldConfig_PAGINATION,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.totalCount"},
			FieldType:     proto.Type_TYPE_INT,
			DisplayName:   "Total count",
			Visible:       false,
			Scope:         toolsproto.ResponseFieldConfig_PAGINATION,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.hasNextPage"},
			FieldType:     proto.Type_TYPE_BOOL,
			DisplayName:   "Has next page",
			Visible:       false,
			Scope:         toolsproto.ResponseFieldConfig_PAGINATION,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.startCursor"},
			FieldType:     proto.Type_TYPE_STRING,
			DisplayName:   "Start cursor",
			Visible:       false,
			Scope:         toolsproto.ResponseFieldConfig_PAGINATION,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.endCursor"},
			FieldType:     proto.Type_TYPE_STRING,
			DisplayName:   "End cursor",
			Visible:       false,
			Scope:         toolsproto.ResponseFieldConfig_PAGINATION,
		},
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.pageInfo.pageNumber"},
			FieldType:     proto.Type_TYPE_INT,
			DisplayName:   "Page Number",
			Visible:       false,
			Scope:         toolsproto.ResponseFieldConfig_PAGINATION,
		},
	}
}

// getPageInfoResponses will return the responses for resultInfo if applicable.
func getResultInfoResponses(schema *proto.Schema, action *proto.Action) ([]*toolsproto.ResponseFieldConfig, error) {
	config := []*toolsproto.ResponseFieldConfig{
		{
			FieldLocation: &toolsproto.JsonPath{Path: "$.resultInfo"},
			FieldType:     proto.Type_TYPE_OBJECT,
			DisplayName:   "ResultInfo",
			Visible:       false,
			Scope:         toolsproto.ResponseFieldConfig_FACETS,
		},
	}

	facetFields := proto.FacetFields(schema, action)
	for _, field := range facetFields {
		config = append(config,
			&toolsproto.ResponseFieldConfig{
				FieldLocation: &toolsproto.JsonPath{Path: fmt.Sprintf("$.resultInfo.%s", field.GetName())},
				FieldType:     proto.Type_TYPE_OBJECT,
				DisplayName:   fmt.Sprintf("%s facets", field.GetName()),
				Visible:       false,
				Scope:         toolsproto.ResponseFieldConfig_FACETS,
			},
		)

		switch field.GetType().GetType() {
		case proto.Type_TYPE_DECIMAL, proto.Type_TYPE_INT:
			config = append(config,
				&toolsproto.ResponseFieldConfig{
					FieldLocation: &toolsproto.JsonPath{Path: fmt.Sprintf("$.resultInfo.%s.min", field.GetName())},
					FieldType:     field.GetType().GetType(),
					DisplayName:   "Minimum",
					Visible:       false,
					Scope:         toolsproto.ResponseFieldConfig_FACETS,
				},
				&toolsproto.ResponseFieldConfig{
					FieldLocation: &toolsproto.JsonPath{Path: fmt.Sprintf("$.resultInfo.%s.max", field.GetName())},
					FieldType:     field.GetType().GetType(),
					DisplayName:   "Maximum",
					Visible:       false,
					Scope:         toolsproto.ResponseFieldConfig_FACETS,
				},
				&toolsproto.ResponseFieldConfig{
					FieldLocation: &toolsproto.JsonPath{Path: fmt.Sprintf("$.resultInfo.%s.avg", field.GetName())},
					FieldType:     proto.Type_TYPE_DECIMAL,
					DisplayName:   "Average",
					Visible:       false,
					Scope:         toolsproto.ResponseFieldConfig_FACETS,
				},
			)
		case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP, proto.Type_TYPE_DURATION:
			config = append(config,
				&toolsproto.ResponseFieldConfig{
					FieldLocation: &toolsproto.JsonPath{Path: fmt.Sprintf("$.resultInfo.%s.min", field.GetName())},
					FieldType:     field.GetType().GetType(),
					DisplayName:   "Minimum",
					Visible:       false,
					Scope:         toolsproto.ResponseFieldConfig_FACETS,
				},
				&toolsproto.ResponseFieldConfig{
					FieldLocation: &toolsproto.JsonPath{Path: fmt.Sprintf("$.resultInfo.%s.max", field.GetName())},
					FieldType:     field.GetType().GetType(),
					DisplayName:   "Maximum",
					Visible:       false,
					Scope:         toolsproto.ResponseFieldConfig_FACETS,
				},
			)
		case proto.Type_TYPE_ENUM, proto.Type_TYPE_STRING:
			config = append(config,
				&toolsproto.ResponseFieldConfig{
					FieldLocation: &toolsproto.JsonPath{Path: fmt.Sprintf("$.resultInfo.%s[*].value", field.GetName())},
					FieldType:     proto.Type_TYPE_STRING,
					DisplayName:   "Value",
					Visible:       false,
					Scope:         toolsproto.ResponseFieldConfig_FACETS,
				},
				&toolsproto.ResponseFieldConfig{
					FieldLocation: &toolsproto.JsonPath{Path: fmt.Sprintf("$.resultInfo.%s[*].count", field.GetName())},
					FieldType:     proto.Type_TYPE_INT,
					DisplayName:   "Count",
					Visible:       false,
					Scope:         toolsproto.ResponseFieldConfig_FACETS,
				},
			)
		default:
			return nil, fmt.Errorf("unsupported facet field type: %s", field.GetType().GetType())
		}
	}

	return config, nil
}

// inferInputType will infer the request field type based on the field name and path prefix. InputMessages are generated
// with some hardcoded field names for certain actions (see makeproto.go):
// Hardcoded field names from schema generation:
// - where
// - first
// - after
// - last
// - before
// - orderBy.
func inferInputType(actionType proto.ActionType, fieldName string) toolsproto.RequestFieldConfig_ScopeType {
	switch actionType {
	case proto.ActionType_ACTION_TYPE_LIST:
		switch fieldName {
		case "where":
			return toolsproto.RequestFieldConfig_FILTERS
		case "orderBy":
			return toolsproto.RequestFieldConfig_SORTING
		case "first", "after", "last", "before", "limit", "offset":
			return toolsproto.RequestFieldConfig_PAGINATION
		}
	case proto.ActionType_ACTION_TYPE_UPDATE:
		switch fieldName {
		case "where":
			return toolsproto.RequestFieldConfig_FILTERS
		case "orderBy":
			return toolsproto.RequestFieldConfig_SORTING
		}
	}

	return toolsproto.RequestFieldConfig_DEFAULT
}

// inferFlowRelatedModel will find a related model to a flow. this is done by looking at the flow inputs
// i.e. if a flow has a model input.
func inferFlowRelatedModel(schema *proto.Schema, flow *proto.Flow) *proto.Model {
	if schema == nil || flow == nil {
		return nil
	}

	mInputs := schema.GetFlowModelInputs(flow)
	if len(mInputs) == 0 {
		return nil
	}

	requiredModels := 0
	relatedModel := ""

	for mName, required := range mInputs {
		if required {
			requiredModels++
			relatedModel = mName

			continue
		}

		if relatedModel == "" {
			relatedModel = mName
		}
	}

	// if there are multiple required model types OR no required types but multiple optional model types we can't associate
	// this flow to a particular model
	if requiredModels != 1 && len(mInputs) > 1 {
		return nil
	}

	return schema.FindModel(relatedModel)
}

// hasOnlyIDInput checks if the tool takes only one input, an ID.
func (t *Tool) hasOnlyIDInput() bool {
	if len(t.ActionConfig.GetInputs()) != 1 {
		return false
	}
	for _, input := range t.ActionConfig.GetInputs() {
		if input.GetFieldType() != proto.Type_TYPE_ID {
			return false
		}
	}

	return true
}

// getIDInputFieldPath returns the path of the first input field that's an ID.
func (t *Tool) getIDInputFieldPath() string {
	for _, input := range t.ActionConfig.GetInputs() {
		if input.GetFieldType() == proto.Type_TYPE_ID && input.GetDisplayName() == casing.ToSentenceCase(fieldNameID) {
			return input.GetFieldLocation().GetPath()
		}
	}

	return ""
}

// getInput finds and returns an inpub by it's path; returns nil if not found.
func (t *Tool) getInput(path string) *toolsproto.RequestFieldConfig {
	for _, input := range t.ActionConfig.GetInputs() {
		if input.GetFieldLocation().GetPath() == path {
			return input
		}
	}

	return nil
}

// getIDResponseFieldPath returns the path of the first response field that's an ID at top level (i.e. results[*].id
// rather than results[*].embedded.id for list actions). Returns empty string if ID is not part of the response.
func (t *Tool) getIDResponseFieldPath() string {
	expectedPath := "$.id"
	if t.Action.IsList() {
		expectedPath = "$.results[*].id"
	}
	for _, response := range t.ActionConfig.GetResponse() {
		if response.GetFieldType() == proto.Type_TYPE_ID && response.GetFieldLocation().GetPath() == expectedPath {
			return response.GetFieldLocation().GetPath()
		}
	}

	return ""
}
