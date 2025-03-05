package tools

import (
	"encoding/json"

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

type ToolConfig struct {
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
}

func (cfg *ToolConfig) applyOn(tool *toolsproto.ActionConfig) {
	if cfg.Name != nil {
		tool.Name = *cfg.Name
	}
	if cfg.Title != nil {
		tool.Title = &toolsproto.StringTemplate{Template: *cfg.Title}
	}
	if cfg.HelpText != nil {
		tool.HelpText = &toolsproto.StringTemplate{Template: *cfg.HelpText}
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
			Label:            &toolsproto.StringTemplate{Template: el.Label},
			Href:             &toolsproto.StringTemplate{Template: el.Href},
			Icon:             el.Icon,
			DisplayOrder:     el.DisplayOrder,
			VisibleCondition: el.VisibleCondition,
		})
	}
	for _, s := range cfg.Sections {
		tool.Sections = append(tool.Sections, &toolsproto.Section{
			Name:  s.Name,
			Title: &toolsproto.StringTemplate{Template: s.Title},
			Description: func() *toolsproto.StringTemplate {
				if s.Description != nil {
					return &toolsproto.StringTemplate{Template: *s.Description}
				}
				return nil
			}(),
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
	if cfg.EntryActivityActions != nil && len(cfg.EntryActivityActions) > 0 {
		tool.EntryActivityActions = cfg.EntryActivityActions.applyOn(tool.EntryActivityActions)
	}
	if cfg.RelatedActions != nil && len(cfg.RelatedActions) > 0 {
		tool.RelatedActions = cfg.RelatedActions.applyOn(tool.RelatedActions)
	}
	if cfg.DisplayLayout != nil {
		tool.DisplayLayout = cfg.DisplayLayout.toProto()
	}
	if cfg.EmbeddedTools != nil && len(cfg.EmbeddedTools) > 0 {
		tool.EmbeddedTools = cfg.EmbeddedTools.applyOn(tool.EmbeddedTools)
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
		input.HelpText = &toolsproto.StringTemplate{Template: *cfg.HelpText}
	}
	if cfg.Locked != nil {
		input.Locked = *cfg.Locked
	}
	if cfg.Placeholder != nil {
		input.Placeholder = &toolsproto.StringTemplate{Template: *cfg.Placeholder}
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
		response.HelpText = &toolsproto.StringTemplate{Template: *cfg.HelpText}
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
	if caps != nil {
		for cap, set := range caps {
			switch cap {
			case CapabilityAudit:
				tool.Capabilities.Audit = set
			case CapabilityComments:
				tool.Capabilities.Comments = set
			}
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

func (cfgs LinkConfigs) applyOn(links []*toolsproto.ActionLink) []*toolsproto.ActionLink {
	newLinks := []*toolsproto.ActionLink{}

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

func (cfg *LinkConfig) applyOn(link *toolsproto.ActionLink) *toolsproto.ActionLink {
	if cfg.isDeleted() {
		return nil
	}
	// we've added a link
	if link == nil {
		return &toolsproto.ActionLink{
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

	return link
}

type ToolGroupConfig struct {
	ID           string      `json:"id,omitempty"`
	Deleted      *bool       `json:"deleted,omitempty"` // if the generated toolgroup has been deleted
	Title        *string     `json:"title,omitempty"`
	DisplayOrder *int32      `json:"display_order,omitempty"`
	Visible      *bool       `json:"visible,omitempty"`
	Tools        LinkConfigs `json:"tools,omitempty"`
	// TODO: ResponseOverrides
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
	// we've added a link
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
				for _, linkCfg := range cfg.Tools {
					tools = append(tools, &toolsproto.ToolGroup_GroupActionLink{
						ActionLink: linkCfg.applyOn(nil),
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

	if cfg.Tools != nil && len(cfg.Tools) > 0 {
		genLinks := []*toolsproto.ActionLink{}
		for _, t := range group.GetTools() {
			genLinks = append(genLinks, t.GetActionLink())
		}

		newLinks := cfg.Tools.applyOn(genLinks)
		group.Tools = []*toolsproto.ToolGroup_GroupActionLink{}
		for _, l := range newLinks {
			group.Tools = append(group.Tools, &toolsproto.ToolGroup_GroupActionLink{
				ActionLink: l,
				// TODO: ResponseOverrides
			})
		}
	}

	return group
}

func extractConfig(generated, updated *toolsproto.ActionConfig) *ToolConfig {
	cfg := &ToolConfig{
		ID:         updated.Id,
		ActionName: updated.ActionName,
	}
	if updated.GetName() != generated.GetName() {
		cfg.Name = &updated.Name
	}
	if updated.GetIcon() != generated.GetIcon() {
		cfg.Icon = updated.Icon
	}
	if change := generated.GetTitle().Diff(updated.GetTitle()); change != "" {
		cfg.Title = &change
	}
	if change := generated.GetHelpText().Diff(updated.GetHelpText()); change != "" {
		cfg.HelpText = &change
	}

	if updated.GetEntitySingle() != generated.GetEntitySingle() {
		cfg.EntitySingle = &updated.EntitySingle
	}
	if updated.GetEntityPlural() != generated.GetEntityPlural() {
		cfg.EntityPlural = &updated.EntityPlural
	}
	if caps := generated.GetCapabilities().Diff(updated.GetCapabilities()); len(caps) > 0 {
		cfg.Capabilities = Capabilities{}
		for cap, set := range caps {
			cfg.Capabilities[Capability(cap)] = set
		}
	}

	cfg.Inputs = InputConfigs{}
	for _, updatedInput := range updated.GetInputs() {
		genInput := generated.FindInput(updatedInput.GetFieldLocation())
		if genInput == nil {
			continue
		}

		// if we have any input changes, set it on the map
		if inputCfg := extractInputConfig(genInput, updatedInput); inputCfg != nil {
			cfg.Inputs[updatedInput.GetFieldLocation().GetPath()] = *inputCfg
		}
	}
	cfg.Response = ResponseConfigs{}
	for _, updatedResponse := range updated.GetResponse() {
		genResponse := generated.FindResponse(updatedResponse.GetFieldLocation())
		if genResponse == nil {
			continue
		}

		// if we have any response changes, set it on the map
		if respCfg := extractResponseConfig(genResponse, updatedResponse); respCfg != nil {
			cfg.Response[updatedResponse.GetFieldLocation().GetPath()] = *respCfg
		}
	}
	cfg.ExternalLinks = ExternalLinks{}
	for _, el := range updated.GetExternalLinks() {
		cfg.ExternalLinks = append(cfg.ExternalLinks, &ExternalLink{
			Label:            el.GetLabel().Template,
			Href:             el.GetHref().Template,
			Icon:             el.Icon,
			DisplayOrder:     el.DisplayOrder,
			VisibleCondition: el.VisibleCondition,
		})
	}
	cfg.Sections = Sections{}
	for _, s := range updated.GetSections() {
		cfg.Sections = append(cfg.Sections, &Section{
			Name:  s.GetName(),
			Title: s.GetTitle().Template,
			Description: func() *string {
				if s.GetDescription() != nil {
					return &s.GetDescription().Template
				}

				return nil
			}(),
			VisibleCondition: s.VisibleCondition,
			DisplayOrder:     s.DisplayOrder,
			Visible:          s.Visible,
		})
	}

	if linkCfg := extractLinkConfig(generated.CreateEntryAction, updated.CreateEntryAction); linkCfg != nil {
		cfg.CreateEntryAction = linkCfg
	}
	if linkCfg := extractLinkConfig(generated.GetEntryAction, updated.GetEntryAction); linkCfg != nil {
		cfg.GetEntryAction = linkCfg
	}
	if linkCfgs := extractLinkConfigs(generated.EntryActivityActions, updated.EntryActivityActions); len(linkCfgs) > 0 {
		cfg.EntryActivityActions = linkCfgs
	}
	if linkCfgs := extractLinkConfigs(generated.RelatedActions, updated.RelatedActions); len(linkCfgs) > 0 {
		cfg.RelatedActions = linkCfgs
	}

	if generated.DisplayLayout.JSON() != updated.DisplayLayout.JSON() {
		cfg.DisplayLayout = &DisplayLayoutConfig{
			Config: updated.DisplayLayout.AsObj(),
		}
	}

	if toolGroupConfigs := extractToolGroupConfigs(generated.EmbeddedTools, updated.EmbeddedTools); len(toolGroupConfigs) > 0 {
		cfg.EmbeddedTools = toolGroupConfigs
	}

	return cfg
}

func extractInputConfig(generated, updated *toolsproto.RequestFieldConfig) *InputConfig {
	cfg := InputConfig{}
	if generated.DisplayName != updated.DisplayName {
		cfg.DisplayName = &updated.DisplayName
	}
	if generated.DisplayOrder != updated.DisplayOrder {
		cfg.DisplayOrder = &updated.DisplayOrder
	}
	if generated.Visible != updated.Visible {
		cfg.Visible = &updated.Visible
	}
	if generated.Locked != updated.Locked {
		cfg.Locked = &updated.Locked
	}
	if change := generated.GetHelpText().Diff(updated.GetHelpText()); change != "" {
		cfg.HelpText = &change
	}
	if change := generated.GetPlaceholder().Diff(updated.GetPlaceholder()); change != "" {
		cfg.Placeholder = &change
	}
	if generated.GetVisibleCondition() != updated.GetVisibleCondition() {
		cfg.VisibleCondition = updated.VisibleCondition
	}
	if generated.GetSectionName() != updated.GetSectionName() {
		cfg.SectionName = updated.SectionName
	}
	if updated.DefaultValue != nil {
		switch updated.DefaultValue.Value.(type) {
		case *toolsproto.ScalarValue_Bool:
			val := updated.DefaultValue.GetBool()
			cfg.DefaultValue = &ScalarValue{BoolValue: &val}
		case *toolsproto.ScalarValue_Float:
			val := updated.DefaultValue.GetFloat()
			cfg.DefaultValue = &ScalarValue{FloatValue: &val}
		case *toolsproto.ScalarValue_String_:
			val := updated.DefaultValue.GetString_()
			cfg.DefaultValue = &ScalarValue{StringValue: &val}
		case *toolsproto.ScalarValue_Integer:
			val := updated.DefaultValue.GetInteger()
			cfg.DefaultValue = &ScalarValue{IntValue: &val}
		}
	}

	if linkCfg := extractLinkConfig(generated.LookupAction, updated.LookupAction); linkCfg != nil {
		cfg.LookupAction = linkCfg
	}
	if linkCfg := extractLinkConfig(generated.GetEntryAction, updated.GetEntryAction); linkCfg != nil {
		cfg.GetEntryAction = linkCfg
	}

	if !cfg.hasChanges() {
		return nil
	}

	return &cfg
}

func extractResponseConfig(generated, updated *toolsproto.ResponseFieldConfig) *ResponseConfig {
	cfg := ResponseConfig{}
	if generated.DisplayName != updated.DisplayName {
		cfg.DisplayName = &updated.DisplayName
	}
	if generated.DisplayOrder != updated.DisplayOrder {
		cfg.DisplayOrder = &updated.DisplayOrder
	}
	if generated.Visible != updated.Visible {
		cfg.Visible = &updated.Visible
	}
	if generated.ImagePreview != updated.ImagePreview {
		cfg.ImagePreview = &updated.ImagePreview
	}
	if change := generated.GetHelpText().Diff(updated.GetHelpText()); change != "" {
		cfg.HelpText = &change
	}
	if generated.GetVisibleCondition() != updated.GetVisibleCondition() {
		cfg.VisibleCondition = updated.VisibleCondition
	}
	if generated.GetSectionName() != updated.GetSectionName() {
		cfg.SectionName = updated.SectionName
	}
	if linkCfg := extractLinkConfig(generated.Link, updated.Link); linkCfg != nil {
		cfg.Link = linkCfg
	}

	if !cfg.hasChanges() {
		return nil
	}

	return &cfg
}

func extractLinkConfigs(generated, updated []*toolsproto.ActionLink) LinkConfigs {
	cfgs := LinkConfigs{}
	// we will use a map to keep track of the generated links that have already been configured
	availableLinks := map[string]bool{}
	for _, l := range generated {
		availableLinks[l.ToolId] = true
	}

	for _, l := range updated {
		if available, ok := availableLinks[l.ToolId]; ok && available {
			// if we have a generated link that hasn't been configured yet (still available), let's use that and diff it with our updated link
			if cfg := extractLinkConfig(toolsproto.FindLinkByToolID(generated, l.ToolId), l); cfg != nil {
				cfgs = append(cfgs, cfg)
			}
			// mark it as used
			availableLinks[l.ToolId] = false

			continue
		}

		// we don't have an available link, so we will add new ones
		cfgs = append(cfgs, extractLinkConfig(nil, l))
	}

	// if there are any generated links that are still available, it means we've removed them, so let's add config for that as well
	for id, available := range availableLinks {
		if available {
			cfgs = append(cfgs, extractLinkConfig(toolsproto.FindLinkByToolID(generated, id), nil))
		}
	}

	if len(cfgs) > 0 {
		return cfgs
	}

	return nil
}

func extractLinkConfig(generated, updated *toolsproto.ActionLink) *LinkConfig {
	// we don't have a link and we didn't add a link
	if generated == nil && updated == nil {
		return nil
	}
	// we have a link and we've removed it
	if generated != nil && updated == nil {
		return &LinkConfig{
			ToolID:  generated.ToolId,
			Deleted: boolPointer(true),
		}
	}

	// we didn't have a link, and now we've added it
	if generated == nil && updated != nil {
		return &LinkConfig{
			ToolID:   updated.ToolId,
			AsDialog: updated.AsDialog,
			Title: func() *string {
				if updated.GetTitle() != nil {
					return &updated.GetTitle().Template
				}
				return nil
			}(),
			Description: func() *string {
				if updated.GetDescription() != nil {
					return &updated.GetDescription().Template
				}
				return nil
			}(),
			DisplayOrder:     &updated.DisplayOrder,
			VisibleCondition: updated.VisibleCondition,
			DataMapping:      updated.GetObjDataMapping(),
		}
	}

	// we may have updated the link config
	cfg := LinkConfig{
		ToolID: generated.ToolId,
	}
	if generated.GetAsDialog() != updated.GetAsDialog() {
		cfg.AsDialog = updated.AsDialog
	}

	if change := generated.GetTitle().Diff(updated.GetTitle()); change != "" {
		cfg.Title = &change
	}
	if change := generated.GetDescription().Diff(updated.GetDescription()); change != "" {
		cfg.Description = &change
	}
	if generated.GetVisibleCondition() != updated.GetVisibleCondition() {
		cfg.VisibleCondition = updated.VisibleCondition
	}
	if generated.DisplayOrder != updated.DisplayOrder {
		cfg.DisplayOrder = &updated.DisplayOrder
	}
	if generated.GetJSONDataMapping() != updated.GetJSONDataMapping() {
		cfg.DataMapping = updated.GetObjDataMapping()
	}

	if !cfg.hasChanges() {
		return nil
	}

	return &cfg
}

func extractToolGroupConfigs(generated, updated []*toolsproto.ToolGroup) ToolGroupConfigs {
	cfgs := ToolGroupConfigs{}

	// we will use a map to keep track of the generated tool groups that have already been configured
	availableGroups := map[string]bool{}
	for _, l := range generated {
		availableGroups[l.GetId()] = true
	}

	for _, g := range updated {
		if available, ok := availableGroups[g.GetId()]; ok && available {
			// if we have a generated group that hasn't been configured yet (still available), let's use that and diff it with our updated group
			if cfg := extractToolGroupConfig(toolsproto.FindToolGroupByID(generated, g.Id), g); cfg != nil {
				cfgs = append(cfgs, cfg)
			}
			// mark it as used
			availableGroups[g.Id] = false

			continue
		}

		// we don't have an available generated group, so we will add a new one
		cfgs = append(cfgs, extractToolGroupConfig(nil, g))
	}

	// if there are any generated groups that are still available, it means we've removed them, so let's add config for that as well
	for id, available := range availableGroups {
		if available {
			cfgs = append(cfgs, extractToolGroupConfig(toolsproto.FindToolGroupByID(generated, id), nil))
		}
	}

	if len(cfgs) > 0 {
		return cfgs
	}

	return nil
}

func extractToolGroupConfig(generated, updated *toolsproto.ToolGroup) *ToolGroupConfig {
	// we don't have a group and we didn't add a group
	if generated == nil && updated == nil {
		return nil
	}
	// we have a group and we've removed it
	if generated != nil && updated == nil {
		return &ToolGroupConfig{
			ID:      generated.GetId(),
			Deleted: boolPointer(true),
		}
	}

	updatedToolLinks := []*toolsproto.ActionLink{}
	for _, groupLink := range updated.Tools {
		updatedToolLinks = append(updatedToolLinks, groupLink.GetActionLink())
	}

	// we didn't have a group, and now we've added it
	if generated == nil && updated != nil {
		return &ToolGroupConfig{
			ID:           updated.Id,
			Visible:      &updated.Visible,
			Tools:        extractLinkConfigs(nil, updatedToolLinks),
			DisplayOrder: &updated.DisplayOrder,
			Title: func() *string {
				if updated.GetTitle() != nil {
					return &updated.GetTitle().Template
				}
				return nil
			}(),
		}
	}

	// we may have updated the tool group config
	generatedToolLinks := []*toolsproto.ActionLink{}
	for _, groupLink := range generated.Tools {
		generatedToolLinks = append(generatedToolLinks, groupLink.GetActionLink())
	}

	cfg := ToolGroupConfig{
		ID: generated.Id,
	}
	if change := generated.GetTitle().Diff(updated.GetTitle()); change != "" {
		cfg.Title = &change
	}
	if generated.DisplayOrder != updated.DisplayOrder {
		cfg.DisplayOrder = &updated.DisplayOrder
	}
	if linkConfigs := extractLinkConfigs(generatedToolLinks, updatedToolLinks); len(linkConfigs) > 0 {
		cfg.Tools = linkConfigs
	}
	if generated.Visible != updated.Visible {
		cfg.Visible = &updated.Visible
	}

	if !cfg.hasChanges() {
		return nil
	}

	return &cfg
}

func boolPointer(v bool) *bool {
	return &v
}

func makeStringTemplate(tmpl *string) *toolsproto.StringTemplate {
	if tmpl == nil {
		return nil
	}
	return &toolsproto.StringTemplate{Template: *tmpl}
}
