package tools

import (
	"encoding/json"
	"fmt"

	"github.com/teamkeel/keel/casing"
	toolsproto "github.com/teamkeel/keel/tools/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

type ToolConfigs []*ToolConfig

func (cfgs ToolConfigs) findByID(id string) *ToolConfig {
	for _, c := range cfgs {
		if c.ID == id {
			return c
		}
	}

	return nil
}

func (cfgs ToolConfigs) hasID(id string) bool {
	c := cfgs.findByID(id)

	return c != nil
}

func (cfgs ToolConfigs) getIDs() []string {
	ids := []string{}
	for _, c := range cfgs {
		ids = append(ids, c.ID)
	}

	return ids
}

type ToolType string

const (
	ToolTypeAction ToolType = "action"
	ToolTypeFlow   ToolType = "flow"
)

type ToolConfig struct {
	ID           string
	Type         ToolType
	ActionConfig *ActionToolConfig
	FlowConfig   *FlowToolConfig
}

func (c ToolConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.config())
}

// To maintain backwards compatibility with existing config files (before we added support for flows), the marshalling
// and unmarshalling of tool configs is done by just marshalling the underlying configuration (i.e.flow config or action config)
func (c *ToolConfig) UnmarshalJSON(data []byte) error {
	// Unmarshal into a generic map to inspect its contents
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Determine type by presence of unique fields
	switch {
	case raw["action_name"] != nil:
		var cfg ActionToolConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("error parsing ActionToolConfig: %w", err)
		}
		c.ActionConfig = &cfg
		c.Type = ToolTypeAction
		c.ID = cfg.ID

		return nil
	case raw["flow_name"] != nil:
		var cfg FlowToolConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("error parsing FlowToolConfig: %w", err)
		}
		c.FlowConfig = &cfg
		c.Type = ToolTypeFlow
		c.ID = cfg.ID

		return nil
	default:
		return fmt.Errorf("unknown tool config type")
	}
}

// config returns the underlying configuration struct:
//   - in the case of action based tools, it will be a ActionToolConfig
//   - for flow based tools, it will be a FlowToolConfig
func (t *ToolConfig) config() configuration {
	switch t.Type {
	case ToolTypeAction:
		return t.ActionConfig
	case ToolTypeFlow:
		return t.FlowConfig
	default:
		return nil
	}
}

func (t *ToolConfig) hasChanges() bool {
	return t.config().hasChanges()
}

// getOperationName returns the underlying action that powers this tool (either an action name or a flow name)
func (t *ToolConfig) getOperationName() string {
	switch t.Type {
	case ToolTypeAction:
		return t.ActionConfig.ActionName
	case ToolTypeFlow:
		return t.FlowConfig.FlowName
	default:
		return ""
	}
}

func (t *ToolConfig) setID(id string) {
	t.ID = id
	if t.Type == ToolTypeAction {
		t.ActionConfig.ID = id
	} else {
		t.FlowConfig.ID = id
	}
}

func (t *ToolConfig) applyOn(tool *toolsproto.Tool) {
	switch t.Type {
	case ToolTypeAction:
		t.ActionConfig.applyOn(tool.ActionConfig)
	case ToolTypeFlow:
		t.FlowConfig.applyOn(tool.FlowConfig)
	}
}

type configuration interface {
	hasChanges() bool
	isDuplicated() bool
	toToolConfig() *ToolConfig
}

// compile time check that config types implement the required interface
var _ configuration = &FlowToolConfig{}
var _ configuration = &ActionToolConfig{}

type FlowToolConfig struct {
	ID                 string           `json:"id,omitempty"`
	FlowName           string           `json:"flow_name,omitempty"`
	Name               *string          `json:"name,omitempty"`
	HelpText           *string          `json:"help_text,omitempty"`
	CompletionRedirect *LinkConfig      `json:"completion_redirect,omitempty"`
	Inputs             FlowInputConfigs `json:"inputs,omitempty"`
}

func (cfg *FlowToolConfig) toToolConfig() *ToolConfig {
	return &ToolConfig{
		ID:         cfg.ID,
		Type:       ToolTypeFlow,
		FlowConfig: cfg,
	}
}

func (cfg *FlowToolConfig) isDuplicated() bool {
	return cfg.ID != casing.ToKebab(cfg.FlowName)
}

func (cfg *FlowToolConfig) hasChanges() bool {
	return cfg.isDuplicated() ||
		cfg.Name != nil ||
		cfg.HelpText != nil ||
		cfg.CompletionRedirect != nil ||
		len(cfg.Inputs) > 0
}

func (cfg *FlowToolConfig) applyOn(tool *toolsproto.FlowConfig) {
	if cfg.Name != nil {
		tool.Name = *cfg.Name
	}
	if cfg.HelpText != nil {
		tool.HelpText = makeStringTemplate(cfg.HelpText)
	}
	if cfg.CompletionRedirect != nil {
		tool.CompletionRedirect = cfg.CompletionRedirect.applyOn(tool.CompletionRedirect)
	}

	for path, inputCfg := range cfg.Inputs {
		if toolInput := tool.FindInputByPath(path); toolInput != nil {
			inputCfg.applyOn(toolInput)
		}
	}
}

type ActionToolConfig struct {
	ID                   string               `json:"id,omitempty"`
	ActionName           string               `json:"action_name,omitempty"`
	Name                 *string              `json:"name,omitempty"`
	Icon                 *string              `json:"icon,omitempty"`
	Title                *string              `json:"title,omitempty"`
	HelpText             *string              `json:"help_text,omitempty"`
	Capabilities         Capabilities         `json:"capabilities,omitempty"`
	EntitySingle         *string              `json:"entity_single,omitempty"`
	EntityPlural         *string              `json:"entity_plural,omitempty"`
	Inputs               InputConfigs         `json:"inputs,omitempty"`
	Response             ResponseConfigs      `json:"response,omitempty"`
	ExternalLinks        ExternalLinks        `json:"external_links,omitempty"`
	Sections             Sections             `json:"sections,omitempty"`
	GetEntryAction       *LinkConfig          `json:"get_entry_action,omitempty"`
	CreateEntryAction    *LinkConfig          `json:"create_entry_action,omitempty"`
	RelatedActions       LinkConfigs          `json:"related_actions,omitempty"`
	EntryActivityActions LinkConfigs          `json:"entry_activity_actions,omitempty"`
	DisplayLayout        *DisplayLayoutConfig `json:"display_layout,omitempty"`
	EmbeddedTools        ToolGroupConfigs     `json:"embedded_tools,omitempty"`
	FilterConfig         *FilterConfig        `json:"filter_config,omitempty"`
}

func (cfg *ActionToolConfig) toToolConfig() *ToolConfig {
	return &ToolConfig{
		ID:           cfg.ID,
		Type:         ToolTypeAction,
		ActionConfig: cfg,
	}
}

func (cfg *ActionToolConfig) isDuplicated() bool {
	return cfg.ID != casing.ToKebab(cfg.ActionName)
}

func (cfg *ActionToolConfig) hasChanges() bool {
	return cfg.isDuplicated() ||
		cfg.Name != nil ||
		cfg.Icon != nil ||
		cfg.Title != nil ||
		cfg.HelpText != nil ||
		len(cfg.Capabilities) > 0 ||
		cfg.EntitySingle != nil ||
		cfg.EntityPlural != nil ||
		len(cfg.Inputs) > 0 ||
		len(cfg.Response) > 0 ||
		len(cfg.ExternalLinks) > 0 ||
		len(cfg.Sections) > 0 ||
		cfg.GetEntryAction != nil ||
		cfg.CreateEntryAction != nil ||
		len(cfg.RelatedActions) > 0 ||
		len(cfg.EntryActivityActions) > 0 ||
		cfg.DisplayLayout != nil ||
		len(cfg.EmbeddedTools) > 0 ||
		cfg.FilterConfig != nil
}

func (cfg *ActionToolConfig) applyOn(tool *toolsproto.ActionConfig) {
	if cfg.Name != nil {
		tool.Name = *cfg.Name
	}
	if cfg.Title != nil {
		tool.Title = makeStringTemplate(cfg.Title)
	}
	if cfg.HelpText != nil {
		tool.HelpText = makeStringTemplate(cfg.HelpText)
	}
	if cfg.Icon != nil {
		tool.Icon = cfg.Icon
	}
	if cfg.EntitySingle != nil {
		tool.EntitySingle = *cfg.EntitySingle
	}
	if cfg.EntityPlural != nil {
		tool.EntityPlural = *cfg.EntityPlural
	}
	cfg.Capabilities.applyOn(tool)

	for path, inputCfg := range cfg.Inputs {
		if toolInput := tool.FindInputByPath(path); toolInput != nil {
			inputCfg.applyOn(toolInput)
		}
	}
	for path, responseCfg := range cfg.Response {
		if toolResponse := tool.FindResponseByPath(path); toolResponse != nil {
			responseCfg.applyOn(toolResponse)
		}
	}
	for _, el := range cfg.ExternalLinks {
		tool.ExternalLinks = append(tool.ExternalLinks, &toolsproto.ExternalLink{
			Label:            makeStringTemplate(&el.Label),
			Href:             makeStringTemplate(&el.Href),
			Icon:             el.Icon,
			DisplayOrder:     el.DisplayOrder,
			VisibleCondition: el.VisibleCondition,
		})
	}
	for _, s := range cfg.Sections {
		tool.Sections = append(tool.Sections, &toolsproto.Section{
			Name:             s.Name,
			Title:            makeStringTemplate(&s.Title),
			Description:      makeStringTemplate(s.Description),
			DisplayOrder:     s.DisplayOrder,
			VisibleCondition: s.VisibleCondition,
			Visible:          s.Visible,
		})
	}
	if cfg.CreateEntryAction != nil {
		tool.CreateEntryAction = cfg.CreateEntryAction.applyOn(tool.CreateEntryAction)
	}
	if cfg.GetEntryAction != nil {
		tool.GetEntryAction = cfg.GetEntryAction.applyOn(tool.GetEntryAction)
	}
	if len(cfg.EntryActivityActions) > 0 {
		tool.EntryActivityActions = cfg.EntryActivityActions.applyOn(tool.EntryActivityActions)
	}
	if len(cfg.RelatedActions) > 0 {
		tool.RelatedActions = cfg.RelatedActions.applyOn(tool.RelatedActions)
	}
	if cfg.DisplayLayout != nil {
		tool.DisplayLayout = cfg.DisplayLayout.toProto()
	}
	if len(cfg.EmbeddedTools) > 0 {
		tool.EmbeddedTools = cfg.EmbeddedTools.applyOn(tool.EmbeddedTools)
	}
	if cfg.FilterConfig != nil {
		tool.FilterConfig = &toolsproto.FilterConfig{
			QuickSearchField: &toolsproto.JsonPath{Path: *cfg.FilterConfig.QuickSearchField},
		}
	}
}

type DisplayLayoutConfig struct {
	Config any `json:"config,omitempty"`
}

func (dl *DisplayLayoutConfig) toJSON() []byte {
	if d, err := json.Marshal(dl.Config); err == nil {
		return d
	}

	return nil
}

func (dl *DisplayLayoutConfig) toProto() *toolsproto.DisplayLayoutConfig {
	var protoDL toolsproto.DisplayLayoutConfig
	if err := protojson.Unmarshal(dl.toJSON(), &protoDL); err != nil {
		return nil
	}

	return &protoDL
}

type InputConfigs map[string]InputConfig
type InputConfig struct {
	DisplayName      *string      `json:"display_name,omitempty"`
	DisplayOrder     *int32       `json:"display_order,omitempty"`
	Visible          *bool        `json:"visible,omitempty"`
	HelpText         *string      `json:"help_text,omitempty"`
	Locked           *bool        `json:"locked,omitempty"`
	Placeholder      *string      `json:"placeholder,omitempty"`
	VisibleCondition *string      `json:"visible_condition,omitempty"`
	SectionName      *string      `json:"section_name,omitempty"`
	DefaultValue     *ScalarValue `json:"default_value,omitempty"`
	LookupAction     *LinkConfig  `json:"lookup_action,omitempty"`
	GetEntryAction   *LinkConfig  `json:"get_entry_action,omitempty"`
}

func (cfg *InputConfig) applyOn(input *toolsproto.RequestFieldConfig) {
	if cfg.DisplayName != nil {
		input.DisplayName = *cfg.DisplayName
	}
	if cfg.DisplayOrder != nil {
		input.DisplayOrder = *cfg.DisplayOrder
	}
	if cfg.Visible != nil {
		input.Visible = *cfg.Visible
	}
	if cfg.HelpText != nil {
		input.HelpText = makeStringTemplate(cfg.HelpText)
	}
	if cfg.Locked != nil {
		input.Locked = *cfg.Locked
	}
	if cfg.Placeholder != nil {
		input.Placeholder = makeStringTemplate(cfg.Placeholder)
	}
	if cfg.VisibleCondition != nil {
		input.VisibleCondition = cfg.VisibleCondition
	}
	if cfg.SectionName != nil {
		input.SectionName = cfg.SectionName
	}
	if cfg.DefaultValue != nil {
		input.DefaultValue = cfg.DefaultValue.toProto()
	}
	if cfg.LookupAction != nil {
		input.LookupAction = cfg.LookupAction.applyOn(input.LookupAction)
	}
	if cfg.GetEntryAction != nil {
		input.GetEntryAction = cfg.GetEntryAction.applyOn(input.GetEntryAction)
	}
}

func (cfg *InputConfig) hasChanges() bool {
	return cfg.DisplayName != nil ||
		cfg.DisplayOrder != nil ||
		cfg.Visible != nil ||
		cfg.HelpText != nil ||
		cfg.Locked != nil ||
		cfg.Placeholder != nil ||
		cfg.VisibleCondition != nil ||
		cfg.SectionName != nil ||
		cfg.DefaultValue != nil ||
		cfg.LookupAction != nil ||
		cfg.GetEntryAction != nil
}

type FlowInputConfigs map[string]FlowInputConfig
type FlowInputConfig struct {
	DisplayName  *string      `json:"display_name,omitempty"`
	DisplayOrder *int32       `json:"display_order,omitempty"`
	HelpText     *string      `json:"help_text,omitempty"`
	Placeholder  *string      `json:"placeholder,omitempty"`
	DefaultValue *ScalarValue `json:"default_value,omitempty"`
}

func (cfg *FlowInputConfig) applyOn(input *toolsproto.FlowInputConfig) {
	if cfg.DisplayName != nil {
		input.DisplayName = *cfg.DisplayName
	}
	if cfg.DisplayOrder != nil {
		input.DisplayOrder = *cfg.DisplayOrder
	}
	if cfg.HelpText != nil {
		input.HelpText = makeStringTemplate(cfg.HelpText)
	}
	if cfg.Placeholder != nil {
		input.Placeholder = makeStringTemplate(cfg.Placeholder)
	}
	if cfg.DefaultValue != nil {
		input.DefaultValue = cfg.DefaultValue.toProto()
	}
}

func (cfg *FlowInputConfig) hasChanges() bool {
	return cfg.DisplayName != nil ||
		cfg.DisplayOrder != nil ||
		cfg.HelpText != nil ||
		cfg.Placeholder != nil ||
		cfg.DefaultValue != nil
}

type ResponseConfigs map[string]ResponseConfig
type ResponseConfig struct {
	DisplayName      *string     `json:"display_name,omitempty"`
	DisplayOrder     *int32      `json:"display_order,omitempty"`
	Visible          *bool       `json:"visible,omitempty"`
	HelpText         *string     `json:"help_text,omitempty"`
	ImagePreview     *bool       `json:"image_preview,omitempty"`
	VisibleCondition *string     `json:"visible_condition,omitempty"`
	SectionName      *string     `json:"section_name,omitempty"`
	Link             *LinkConfig `json:"link,omitempty"`
}

func (cfg ResponseConfig) applyOn(response *toolsproto.ResponseFieldConfig) {
	if cfg.DisplayName != nil {
		response.DisplayName = *cfg.DisplayName
	}
	if cfg.DisplayOrder != nil {
		response.DisplayOrder = *cfg.DisplayOrder
	}
	if cfg.Visible != nil {
		response.Visible = *cfg.Visible
	}
	if cfg.HelpText != nil {
		response.HelpText = makeStringTemplate(cfg.HelpText)
	}
	if cfg.ImagePreview != nil {
		response.ImagePreview = *cfg.ImagePreview
	}
	if cfg.VisibleCondition != nil {
		response.VisibleCondition = cfg.VisibleCondition
	}
	if cfg.SectionName != nil {
		response.SectionName = cfg.SectionName
	}
	if cfg.Link != nil {
		response.Link = cfg.Link.applyOn(response.Link)
	}
}

func (cfg ResponseConfig) hasChanges() bool {
	return cfg.DisplayName != nil ||
		cfg.DisplayOrder != nil ||
		cfg.Visible != nil ||
		cfg.HelpText != nil ||
		cfg.ImagePreview != nil ||
		cfg.VisibleCondition != nil ||
		cfg.SectionName != nil ||
		cfg.Link != nil
}

type Capability string
type Capabilities map[Capability]bool

const (
	CapabilityAudit    Capability = "audit"
	CapabilityComments Capability = "comments"
)

func (caps Capabilities) applyOn(tool *toolsproto.ActionConfig) {
	for cap, set := range caps {
		switch cap {
		case CapabilityAudit:
			tool.Capabilities.Audit = set
		case CapabilityComments:
			tool.Capabilities.Comments = set
		}
	}
}

type ExternalLink struct {
	Label            string  `json:"label,omitempty"`
	Href             string  `json:"href,omitempty"`
	Icon             *string `json:"icon,omitempty"`
	DisplayOrder     int32   `json:"display_order,omitempty"`
	VisibleCondition *string `json:"visible_condition,omitempty"`
}

type ExternalLinks []*ExternalLink

type Section struct {
	Name             string  `json:"name,omitempty"`
	Title            string  `json:"title,omitempty"`
	Description      *string `json:"description,omitempty"`
	VisibleCondition *string `json:"visible_condition,omitempty"`
	DisplayOrder     int32   `json:"display_order,omitempty"`
	Visible          bool    `json:"visible,omitempty"`
}

type Sections []*Section

type ScalarValue struct {
	StringValue *string  `json:"string_value,omitempty"`
	IntValue    *int32   `json:"int_value,omitempty"`
	FloatValue  *float32 `json:"float_value,omitempty"`
	BoolValue   *bool    `json:"bool_value,omitempty"`
	NullValue   *bool    `json:"null_value,omitempty"`
}

func (v *ScalarValue) toProto() *toolsproto.ScalarValue {
	if v.StringValue != nil {
		return &toolsproto.ScalarValue{Value: &toolsproto.ScalarValue_String_{String_: *v.StringValue}}
	}
	if v.IntValue != nil {
		return &toolsproto.ScalarValue{Value: &toolsproto.ScalarValue_Integer{Integer: *v.IntValue}}
	}
	if v.StringValue != nil {
		return &toolsproto.ScalarValue{Value: &toolsproto.ScalarValue_Float{Float: *v.FloatValue}}
	}
	if v.StringValue != nil {
		return &toolsproto.ScalarValue{Value: &toolsproto.ScalarValue_Bool{Bool: *v.BoolValue}}
	}
	if v.StringValue != nil {
		return &toolsproto.ScalarValue{Value: &toolsproto.ScalarValue_Null{Null: *v.NullValue}}
	}

	return nil
}

type LinkConfig struct {
	ToolID           string  `json:"tool_id"`
	Deleted          *bool   `json:"deleted,omitempty"` // if the generated link has been deleted
	Title            *string `json:"title,omitempty"`
	Description      *string `json:"description,omitempty"`
	AsDialog         *bool   `json:"as_dialog,omitempty"`
	DisplayOrder     *int32  `json:"display_order,omitempty"`
	VisibleCondition *string `json:"visible_condition,omitempty"`
	DataMapping      []any   `json:"data_mapping,omitempty"`
}

type LinkConfigs []*LinkConfig

func (cfgs LinkConfigs) find(toolID string) *LinkConfig {
	for _, tl := range cfgs {
		if tl.ToolID == toolID {
			return tl
		}
	}

	return nil
}

func (cfg LinkConfig) hasChanges() bool {
	return cfg.Deleted != nil ||
		cfg.Title != nil ||
		cfg.Description != nil ||
		cfg.AsDialog != nil ||
		cfg.DisplayOrder != nil ||
		cfg.VisibleCondition != nil ||
		cfg.DataMapping != nil
}

func (cfg *LinkConfig) getDataMapping() []*toolsproto.DataMapping {
	if cfg.DataMapping == nil {
		return nil
	}
	dataMappings := []*toolsproto.DataMapping{}
	for _, d := range cfg.DataMapping {
		jsonStr, err := json.Marshal(d)
		if err != nil {
			return nil
		}

		var dm toolsproto.DataMapping
		if err := protojson.Unmarshal(jsonStr, &dm); err != nil {
			return nil
		}
		dataMappings = append(dataMappings, &dm)
	}
	return dataMappings
}

// isDeleted tells us if the link has been ... deleted
func (cfg *LinkConfig) isDeleted() bool {
	if cfg != nil && cfg.Deleted != nil {
		return *cfg.Deleted
	}

	return false
}

func (cfgs LinkConfigs) applyOn(links []*toolsproto.ToolLink) []*toolsproto.ToolLink {
	newLinks := []*toolsproto.ToolLink{}

	// add all configured links and new links. If links are deleted, they are skipped
	for _, cfg := range cfgs {
		if configured := cfg.applyOn(toolsproto.FindLinkByToolID(links, cfg.ToolID)); configured != nil {
			newLinks = append(newLinks, configured)
		}
	}

	// carry over links that haven't been configured/deleted
	for _, l := range links {
		if cfg := cfgs.find(l.ToolId); cfg == nil {
			newLinks = append(newLinks, l)
		}
	}

	return newLinks
}

func (cfg *LinkConfig) applyOn(link *toolsproto.ToolLink) *toolsproto.ToolLink {
	if cfg == nil {
		return nil
	}
	if cfg.isDeleted() {
		return nil
	}
	// we've added a link
	if link == nil {
		return &toolsproto.ToolLink{
			ToolId:      cfg.ToolID,
			Description: makeStringTemplate(cfg.Description),
			Title:       makeStringTemplate(cfg.Title),
			DisplayOrder: func() int32 {
				if cfg.DisplayOrder != nil {
					return *cfg.DisplayOrder
				}
				return 0
			}(),
			AsDialog:         cfg.AsDialog,
			VisibleCondition: cfg.VisibleCondition,
			Data:             cfg.getDataMapping(),
		}
	}

	if cfg.Title != nil {
		link.Title = makeStringTemplate(cfg.Title)
	}
	if cfg.Description != nil {
		link.Title = makeStringTemplate(cfg.Description)
	}
	if cfg.DisplayOrder != nil {
		link.DisplayOrder = *cfg.DisplayOrder
	}
	if cfg.AsDialog != nil {
		link.AsDialog = cfg.AsDialog
	}
	if cfg.VisibleCondition != nil {
		link.VisibleCondition = cfg.VisibleCondition
	}
	if dm := cfg.getDataMapping(); dm != nil {
		link.Data = dm
	}

	return link
}

type ToolGroupConfig struct {
	ID           string               `json:"id,omitempty"`
	Deleted      *bool                `json:"deleted,omitempty"` // if the generated toolgroup has been deleted
	Title        *string              `json:"title,omitempty"`
	DisplayOrder *int32               `json:"display_order,omitempty"`
	Visible      *bool                `json:"visible,omitempty"`
	Tools        ToolGroupLinkConfigs `json:"tools,omitempty"`
}
type ToolGroupConfigs []*ToolGroupConfig

func (cfg *ToolGroupConfig) hasChanges() bool {
	return cfg.Deleted != nil ||
		cfg.Title != nil ||
		cfg.DisplayOrder != nil ||
		cfg.Visible != nil ||
		cfg.Tools != nil
}

func (cfgs ToolGroupConfigs) find(id string) *ToolGroupConfig {
	for _, tg := range cfgs {
		if tg.ID == id {
			return tg
		}
	}

	return nil
}

func (cfg *ToolGroupConfig) isDeleted() bool {
	if cfg != nil && cfg.Deleted != nil {
		return *cfg.Deleted
	}

	return false
}

func (cfgs ToolGroupConfigs) applyOn(groups []*toolsproto.ToolGroup) []*toolsproto.ToolGroup {
	newTools := []*toolsproto.ToolGroup{}

	// add all configured tool groups and new groups. If groups are deleted, they are skipped
	for _, cfg := range cfgs {
		if configured := cfg.applyOn(toolsproto.FindToolGroupByID(groups, cfg.ID)); configured != nil {
			newTools = append(newTools, configured)
		}
	}

	// carry over groups that haven't been configured/deleted
	for _, g := range groups {
		if cfg := cfgs.find(g.Id); cfg == nil {
			newTools = append(newTools, g)
		}
	}

	return newTools
}

func (cfg *ToolGroupConfig) applyOn(group *toolsproto.ToolGroup) *toolsproto.ToolGroup {
	if cfg.isDeleted() {
		return nil
	}
	// we've added a group
	if group == nil {
		return &toolsproto.ToolGroup{
			Id:    cfg.ID,
			Title: makeStringTemplate(cfg.Title),
			DisplayOrder: func() int32 {
				if cfg.DisplayOrder != nil {
					return *cfg.DisplayOrder
				}
				return 0
			}(),
			Visible: func() bool {
				if cfg.Visible != nil {
					return *cfg.Visible
				}
				return false
			}(),
			Tools: func() []*toolsproto.ToolGroup_GroupActionLink {
				tools := []*toolsproto.ToolGroup_GroupActionLink{}
				for _, toolLink := range cfg.Tools {
					tools = append(tools, &toolsproto.ToolGroup_GroupActionLink{
						ActionLink:        toolLink.ActionLink.applyOn(nil),
						ResponseOverrides: toolLink.getResponseOverrides(),
					})
				}
				return tools
			}(),
		}
	}

	if cfg.Title != nil {
		group.Title = makeStringTemplate(cfg.Title)
	}
	if cfg.DisplayOrder != nil {
		group.DisplayOrder = *cfg.DisplayOrder
	}
	if cfg.Visible != nil {
		group.Visible = *cfg.Visible
	}

	if len(cfg.Tools) > 0 {
		group.Tools = cfg.Tools.applyOn(group.Tools)
	}

	return group
}

type ToolGroupLinkConfig struct {
	Deleted           *bool           `json:"deleted,omitempty"`
	ActionLink        *LinkConfig     `json:"action_link,omitempty"`
	ResponseOverrides map[string]bool `json:"response_overrides,omitempty"`
}

type ToolGroupLinkConfigs []*ToolGroupLinkConfig

func (cfg *ToolGroupLinkConfig) hasChanges() bool {
	return len(cfg.ResponseOverrides) > 0 || cfg.isDeleted() || (cfg.ActionLink != nil && cfg.ActionLink.hasChanges())
}

func (cfg *ToolGroupLinkConfig) getToolLinkID() string {
	if cfg == nil || cfg.ActionLink == nil {
		return ""
	}

	return cfg.ActionLink.ToolID
}

func (cfgs ToolGroupLinkConfigs) applyOn(links []*toolsproto.ToolGroup_GroupActionLink) []*toolsproto.ToolGroup_GroupActionLink {
	newTools := []*toolsproto.ToolGroup_GroupActionLink{}

	// add all configured links and new links. If links are deleted, they are skipped
	for _, cfg := range cfgs {
		if cfg.getToolLinkID() == "" {
			// skip embedding configs with invalid tools links
			continue
		}
		if configured := cfg.applyOn(toolsproto.FindToolGroupLinkByToolID(links, cfg.getToolLinkID())); configured != nil {
			newTools = append(newTools, configured)
		}
	}

	// carry over links that haven't been configured/deleted
	for _, l := range links {
		if cfg := cfgs.find(l.ActionLink.ToolId); cfg == nil {
			newTools = append(newTools, l)
		}
	}

	return newTools
}

func (cfgs ToolGroupLinkConfigs) find(toolID string) *ToolGroupLinkConfig {
	for _, tl := range cfgs {
		if tl.getToolLinkID() == toolID {
			return tl
		}
	}

	return nil
}
func (cfg *ToolGroupLinkConfig) isDeleted() bool {
	if cfg.Deleted != nil {
		return *cfg.Deleted
	}

	return false
}

func (cfg *ToolGroupLinkConfig) getResponseOverrides() []*toolsproto.ResponseOverrides {
	if cfg == nil {
		return nil
	}

	ret := []*toolsproto.ResponseOverrides{}
	for path, visible := range cfg.ResponseOverrides {
		ret = append(ret, &toolsproto.ResponseOverrides{
			FieldLocation: &toolsproto.JsonPath{Path: path},
			Visible:       visible,
		})
	}
	return ret
}

func (cfg *ToolGroupLinkConfig) applyOn(link *toolsproto.ToolGroup_GroupActionLink) *toolsproto.ToolGroup_GroupActionLink {
	if cfg == nil || cfg.isDeleted() {
		return nil
	}

	// we've added a tool group link
	if link == nil {
		return &toolsproto.ToolGroup_GroupActionLink{
			ActionLink:        cfg.ActionLink.applyOn(nil),
			ResponseOverrides: cfg.getResponseOverrides(),
		}
	}

	// we've updated a tool group link
	link.ActionLink = cfg.ActionLink.applyOn(link.ActionLink)
	link.ResponseOverrides = cfg.getResponseOverrides()

	return link
}

type FilterConfig struct {
	QuickSearchField *string `json:"quick_search_field,omitempty"`
}

func makeStringTemplate(tmpl *string) *toolsproto.StringTemplate {
	if tmpl != nil {
		return &toolsproto.StringTemplate{Template: *tmpl}
	}

	return nil
}
