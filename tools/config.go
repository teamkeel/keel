package tools

import (
	toolsproto "github.com/teamkeel/keel/tools/proto"
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
	ID            string          `json:"id,omitempty"`
	ActionName    string          `json:"action_name,omitempty"`
	Name          *string         `json:"name,omitempty"`
	Icon          *string         `json:"icon,omitempty"`
	Title         *string         `json:"title,omitempty"`
	HelpText      *string         `json:"help_text,omitempty"`
	Capabilities  Capabilities    `json:"capabilities,omitempty"`
	EntitySingle  *string         `json:"entity_single,omitempty"`
	EntityPlural  *string         `json:"entity_plural,omitempty"`
	Inputs        InputConfigs    `json:"inputs,omitempty"`
	Response      ResponseConfigs `json:"response,omitempty"`
	ExternalLinks ExternalLinks   `json:"external_links,omitempty"`
	// TODO: RelatedActions       []*ToolLink
	// TODO: EntryActivityActions []*ToolLink
	// TODO: GetEntryAction       *ToolLink
	// TODO: CreateEntryAction    *ToolLink
	// TODO: EmbeddedTools        ToolGroups
	// TODO: DisplayLayout        *DisplayLayout
	// TODO: Sections             Sections
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
}

type InputConfigs map[string]InputConfig
type InputConfig struct {
	DisplayName      *string `json:"display_name,omitempty"`
	DisplayOrder     *int32  `json:"display_order,omitempty"`
	Visible          *bool   `json:"visible,omitempty"`
	HelpText         *string `json:"help_text,omitempty"`
	Locked           *bool   `json:"locked,omitempty"`
	Placeholder      *string `json:"placeholder,omitempty"`
	VisibleCondition *string `json:"visible_condition,omitempty"`
	SectionName      *string `json:"section_name,omitempty"`
	// TODO: DefaultValue
	// TODO Lookup Link
	// TODO GetEntry Link
}

func (cfg InputConfig) applyOn(input *toolsproto.RequestFieldConfig) {
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
}

func (cfg InputConfig) hasChanges() bool {
	return cfg.DisplayName != nil ||
		cfg.DisplayOrder != nil ||
		cfg.Visible != nil ||
		cfg.HelpText != nil ||
		cfg.Locked != nil ||
		cfg.Placeholder != nil ||
		cfg.VisibleCondition != nil ||
		cfg.SectionName != nil
}

type ResponseConfigs map[string]ResponseConfig
type ResponseConfig struct {
	DisplayName      *string `json:"display_name,omitempty"`
	DisplayOrder     *int32  `json:"display_order,omitempty"`
	Visible          *bool   `json:"visible,omitempty"`
	HelpText         *string `json:"help_text,omitempty"`
	ImagePreview     *bool   `json:"image_preview,omitempty"`
	VisibleCondition *string `json:"visible_condition,omitempty"`
	SectionName      *string `json:"section_name,omitempty"`
	// TODO  Link
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
}

func (cfg ResponseConfig) hasChanges() bool {
	return cfg.DisplayName != nil ||
		cfg.DisplayOrder != nil ||
		cfg.Visible != nil ||
		cfg.HelpText != nil ||
		cfg.ImagePreview != nil ||
		cfg.VisibleCondition != nil ||
		cfg.SectionName != nil
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

	if !cfg.hasChanges() {
		return nil
	}

	return &cfg
}
