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
	ID           string       `json:"id,omitempty"`
	ActionName   string       `json:"action_name,omitempty"`
	Name         string       `json:"name,omitempty"`
	Icon         string       `json:"icon,omitempty"`
	Title        string       `json:"title,omitempty"`
	HelpText     string       `json:"help_text,omitempty"`
	Capabilities Capabilities `json:"capabilities,omitempty"`
	EntitySingle string       `json:"entity_single,omitempty"`
	EntityPlural string       `json:"entity_plural,omitempty"`
	Inputs       InputConfigs `json:"inputs,omitempty"`
	// TODO: RelatedActions       []*ToolLink
	// TODO: ExternalLinks        ExternalLinks
	// TODO: EntryActivityActions []*ToolLink
	// TODO: EmbeddedTools        ToolGroups
	// TODO: GetEntryAction       *ToolLink
	// TODO: CreateEntryAction    *ToolLink
	// TODO: DisplayLayout        *DisplayLayout
	// TODO: Sections             Sections
	// TODO: Response
}

func (cfg *ToolConfig) applyOn(tool *toolsproto.ActionConfig) {
	if cfg.Name != "" {
		tool.Name = cfg.Name
	}
	if cfg.Title != "" {
		tool.Title = &toolsproto.StringTemplate{Template: cfg.Title}
	}
	if cfg.HelpText != "" {
		tool.HelpText = &toolsproto.StringTemplate{Template: cfg.HelpText}
	}
	if cfg.Icon != "" {
		tool.Icon = &cfg.Icon
	}
	if cfg.EntitySingle != "" {
		tool.EntitySingle = cfg.EntitySingle
	}
	if cfg.EntityPlural != "" {
		tool.EntityPlural = cfg.EntityPlural
	}
	cfg.Capabilities.applyOn(tool)

	for path, inputCfg := range cfg.Inputs {
		if toolInput := tool.FindInputByPath(path); toolInput != nil {
			inputCfg.applyOn(toolInput)
		}
	}
}

type InputConfigs map[string]InputConfig
type InputConfig struct {
	DisplayName      string `json:"display_name,omitempty"`
	DisplayOrder     int    `json:"display_order,omitempty"`
	Visible          bool   `json:"visible,omitempty"`
	HelpText         string `json:"help_text,omitempty"`
	Locked           bool   `json:"locked,omitempty"`
	Placeholder      string `json:"placeholder,omitempty"`
	VisibleCondition string `json:"visible_condition,omitempty"`
	SectionName      string `json:"section_name,omitempty"`
	// TODO: DefaultValue
	// TODO Lookup Link
	// TODO GetEntry Link
}

func (cfg InputConfig) applyOn(input *toolsproto.RequestFieldConfig) {
	if cfg.DisplayName != "" {
		input.DisplayName = cfg.DisplayName
	}
	if cfg.DisplayOrder != int(input.DisplayOrder) {
		input.DisplayOrder = int32(cfg.DisplayOrder)
	}
	if cfg.Visible != input.Visible {
		input.Visible = cfg.Visible
	}
	if cfg.HelpText != "" {
		input.HelpText = &toolsproto.StringTemplate{Template: cfg.HelpText}
	}
	if cfg.Locked != input.Locked {
		input.Locked = cfg.Locked
	}
	if cfg.Placeholder != "" {
		input.Placeholder = &toolsproto.StringTemplate{Template: cfg.Placeholder}
	}
	if cfg.VisibleCondition != "" {
		input.VisibleCondition = &cfg.VisibleCondition
	}
	if cfg.SectionName != "" {
		input.SectionName = &cfg.SectionName
	}
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

func extractConfig(generated, updated *toolsproto.ActionConfig) *ToolConfig {
	cfg := &ToolConfig{
		ID:         updated.Id,
		ActionName: updated.ActionName,
	}
	if updated.GetName() != generated.GetName() {
		cfg.Name = updated.GetName()
	}
	if updated.GetIcon() != generated.GetIcon() {
		cfg.Icon = updated.GetIcon()
	}

	cfg.Title = generated.GetTitle().Diff(updated.GetTitle())
	cfg.HelpText = generated.GetHelpText().Diff(updated.GetHelpText())

	if updated.GetEntitySingle() != generated.GetEntitySingle() {
		cfg.EntitySingle = updated.GetEntitySingle()
	}
	if updated.GetEntityPlural() != generated.GetEntityPlural() {
		cfg.EntityPlural = updated.GetEntityPlural()
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

	return cfg
}

func extractInputConfig(generated, updated *toolsproto.RequestFieldConfig) *InputConfig {
	cfg := InputConfig{}
	changed := false

	if generated.DisplayName != updated.DisplayName {
		cfg.DisplayName = updated.DisplayName
		changed = true
	}
	if generated.DisplayOrder != updated.DisplayOrder {
		cfg.DisplayOrder = int(updated.DisplayOrder)
		changed = true
	}
	if generated.Visible != updated.Visible {
		cfg.Visible = updated.Visible
		changed = true
	}
	if generated.Locked != updated.Locked {
		cfg.Locked = updated.Locked
		changed = true
	}
	if cfg.HelpText = generated.GetHelpText().Diff(updated.GetHelpText()); cfg.HelpText != "" {
		changed = true
	}
	if cfg.HelpText = generated.GetPlaceholder().Diff(updated.GetPlaceholder()); cfg.Placeholder != "" {
		changed = true
	}
	if generated.GetVisibleCondition() != updated.GetVisibleCondition() {
		cfg.VisibleCondition = updated.GetVisibleCondition()
		changed = true
	}
	if generated.GetSectionName() != updated.GetSectionName() {
		cfg.SectionName = updated.GetSectionName()
		changed = true
	}

	if !changed {
		return nil
	}

	return &cfg
}
