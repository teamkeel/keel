package tools

import (
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

func extractConfig(generated, updated *toolsproto.Tool) *ToolConfig {
	if generated.Type != updated.Type {
		return nil
	}

	var cfg *ToolConfig

	if generated.IsActionBased() {
		cfg = extractActionConfig(generated.ActionConfig, updated.ActionConfig).toToolConfig()
	} else {
		cfg = extractFlowConfig(generated.FlowConfig, updated.FlowConfig).toToolConfig()
	}
	cfg.setID(updated.GetId())
	return cfg
}

func extractFlowConfig(generated, updated *toolsproto.FlowConfig) *FlowToolConfig {
	cfg := &FlowToolConfig{
		FlowName:           updated.FlowName,
		Name:               diffString(generated.GetName(), updated.GetName()),
		HelpText:           diffStringTemplate(generated.GetHelpText(), updated.GetHelpText()),
		CompletionRedirect: extractLinkConfig(generated.CompletionRedirect, updated.CompletionRedirect),
	}

	cfg.Inputs = FlowInputConfigs{}
	for _, updatedInput := range updated.GetInputs() {
		genInput := generated.FindInput(updatedInput.GetFieldLocation())
		if genInput == nil {
			continue
		}

		// if we have any input changes, set it on the map
		if inputCfg := extractFlowInputConfig(genInput, updatedInput); inputCfg != nil {
			cfg.Inputs[updatedInput.GetFieldLocation().GetPath()] = *inputCfg
		}
	}

	return cfg
}

func extractActionConfig(generated, updated *toolsproto.ActionConfig) *ActionToolConfig {
	cfg := &ActionToolConfig{
		ID:                   updated.Id,
		ActionName:           updated.ActionName,
		Name:                 diffString(generated.GetName(), updated.GetName()),
		Icon:                 diffString(generated.GetIcon(), updated.GetIcon()),
		Title:                diffStringTemplate(generated.GetTitle(), updated.GetTitle()),
		HelpText:             diffStringTemplate(generated.GetHelpText(), updated.GetHelpText()),
		EntitySingle:         diffString(generated.GetEntitySingle(), updated.GetEntitySingle()),
		EntityPlural:         diffString(generated.GetEntityPlural(), updated.GetEntityPlural()),
		Pagination:           extractPaginationConfig(generated.Pagination, updated.Pagination),
		CreateEntryAction:    extractLinkConfig(generated.CreateEntryAction, updated.CreateEntryAction),
		GetEntryAction:       extractLinkConfig(generated.GetEntryAction, updated.GetEntryAction),
		EntryActivityActions: extractLinkConfigs(generated.EntryActivityActions, updated.EntryActivityActions),
		RelatedActions:       extractLinkConfigs(generated.RelatedActions, updated.RelatedActions),
		EmbeddedTools:        extractToolGroupConfigs(generated.EmbeddedTools, updated.EmbeddedTools),
		FilterConfig:         extractFilterConfig(generated.FilterConfig, updated.FilterConfig),
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
			Name:             s.GetName(),
			Title:            s.GetTitle().Template,
			Description:      diffStringTemplate(nil, s.GetDescription()),
			VisibleCondition: s.VisibleCondition,
			DisplayOrder:     s.DisplayOrder,
			Visible:          s.Visible,
		})
	}

	if generated.DisplayLayout.JSON() != updated.DisplayLayout.JSON() {
		cfg.DisplayLayout = &DisplayLayoutConfig{
			Config: updated.DisplayLayout.AsObj(),
		}
	}

	return cfg
}

func extractInputConfig(generated, updated *toolsproto.RequestFieldConfig) *InputConfig {
	cfg := InputConfig{
		DisplayName:      diffString(generated.GetDisplayName(), updated.GetDisplayName()),
		DisplayOrder:     diffInt(generated.GetDisplayOrder(), updated.GetDisplayOrder()),
		Visible:          diffBool(generated.GetVisible(), updated.GetVisible()),
		Locked:           diffBool(generated.GetLocked(), updated.GetLocked()),
		HelpText:         diffStringTemplate(generated.GetHelpText(), updated.GetHelpText()),
		Placeholder:      diffStringTemplate(generated.GetPlaceholder(), updated.GetPlaceholder()),
		VisibleCondition: diffString(generated.GetVisibleCondition(), updated.GetVisibleCondition()),
		SectionName:      diffString(generated.GetSectionName(), updated.GetSectionName()),
		LookupAction:     extractLinkConfig(generated.LookupAction, updated.LookupAction),
		GetEntryAction:   extractLinkConfig(generated.GetEntryAction, updated.GetEntryAction),
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

	if !cfg.hasChanges() {
		return nil
	}

	return &cfg
}

func extractFlowInputConfig(generated, updated *toolsproto.FlowInputConfig) *FlowInputConfig {
	cfg := FlowInputConfig{
		DisplayName:  diffString(generated.GetDisplayName(), updated.GetDisplayName()),
		DisplayOrder: diffInt(generated.GetDisplayOrder(), updated.GetDisplayOrder()),
		HelpText:     diffStringTemplate(generated.GetHelpText(), updated.GetHelpText()),
		Placeholder:  diffStringTemplate(generated.GetPlaceholder(), updated.GetPlaceholder()),
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

	if !cfg.hasChanges() {
		return nil
	}

	return &cfg
}

func extractResponseConfig(generated, updated *toolsproto.ResponseFieldConfig) *ResponseConfig {
	cfg := ResponseConfig{
		DisplayName:      diffString(generated.GetDisplayName(), updated.GetDisplayName()),
		DisplayOrder:     diffInt(generated.GetDisplayOrder(), updated.GetDisplayOrder()),
		Visible:          diffBool(generated.GetVisible(), updated.GetVisible()),
		ImagePreview:     diffBool(generated.GetImagePreview(), updated.GetImagePreview()),
		HelpText:         diffStringTemplate(generated.GetHelpText(), updated.GetHelpText()),
		VisibleCondition: diffString(generated.GetVisibleCondition(), updated.GetVisibleCondition()),
		SectionName:      diffString(generated.GetSectionName(), updated.GetSectionName()),
		Link:             extractLinkConfig(generated.Link, updated.Link),
	}

	if !cfg.hasChanges() {
		return nil
	}

	return &cfg
}

func extractLinkConfigs(generated, updated []*toolsproto.ToolLink) LinkConfigs {
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


func extractPaginationConfig(generated, updated *toolsproto.CursorPaginationConfig) *PaginationConfig {
	if generated == nil && updated == nil {
		return nil
	}

	if generated != nil && updated == nil {
		return nil
	}

	if generated == nil && updated != nil {
		return &PaginationConfig{
			PageSize: extractPageSizeConfig(nil, updated.PageSize),
		}
	}

	return &PaginationConfig{
		PageSize: extractPageSizeConfig(generated.PageSize, updated.PageSize),
	}
}

func extractPageSizeConfig(generated, updated *toolsproto.CursorPaginationConfig_PageSizeConfig) *PageSizeConfig {
	if generated == nil && updated == nil {
		return nil
	}

	if generated != nil && updated == nil {
		return nil
	}

	if generated == nil && updated != nil {
		return &PageSizeConfig{
			DefaultValue: &updated.DefaultValue,
		}
	}

	return &PageSizeConfig{
		DefaultValue: diffInt(generated.DefaultValue, updated.DefaultValue),
	}
}

func extractLinkConfig(generated, updated *toolsproto.ToolLink) *LinkConfig {
	// we don't have a link and we didn't add a link
	if generated == nil && updated == nil {
		return nil
	}
	// we have a link and we've removed it
	if generated != nil && updated == nil {
		return &LinkConfig{
			ToolID:  generated.ToolId,
			Deleted: diffBool(false, true),
		}
	}

	// we didn't have a link, and now we've added it
	if generated == nil && updated != nil {
		return &LinkConfig{
			ToolID:           updated.ToolId,
			AsDialog:         updated.AsDialog,
			Title:            diffStringTemplate(nil, updated.GetTitle()),
			Description:      diffStringTemplate(nil, updated.GetDescription()),
			DisplayOrder:     &updated.DisplayOrder,
			VisibleCondition: updated.VisibleCondition,
			DataMapping:      updated.GetObjDataMapping(),
			SkipConfirmation: updated.SkipConfirmation,
			Emphasize:        updated.Emphasize,
		}
	}

	// we may have updated the link config
	cfg := LinkConfig{
		ToolID:           generated.ToolId,
		AsDialog:         diffBool(generated.GetAsDialog(), updated.GetAsDialog()),
		Title:            diffStringTemplate(generated.GetTitle(), updated.GetTitle()),
		Description:      diffStringTemplate(generated.GetDescription(), updated.GetDescription()),
		DisplayOrder:     diffInt(generated.GetDisplayOrder(), updated.GetDisplayOrder()),
		VisibleCondition: diffString(generated.GetVisibleCondition(), updated.GetVisibleCondition()),
		SkipConfirmation: diffBool(generated.GetSkipConfirmation(), updated.GetSkipConfirmation()),
		Emphasize:        diffBool(generated.GetEmphasize(), updated.GetEmphasize()),
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

func extractFilterConfig(generated, updated *toolsproto.FilterConfig) *FilterConfig {
	if generated == nil || updated == nil {
		return nil
	}

	if updated.GetQuickSearchField() != nil {
		return &FilterConfig{
			QuickSearchField: &updated.GetQuickSearchField().Path,
		}
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
			Deleted: diffBool(false, true),
		}
	}

	// we didn't have a group, and now we've added it
	if generated == nil && updated != nil {
		return &ToolGroupConfig{
			ID:           updated.Id,
			Visible:      &updated.Visible,
			Tools:        extractToolGroupLinkConfigs(nil, updated.Tools),
			DisplayOrder: &updated.DisplayOrder,
			Title:        diffStringTemplate(nil, updated.GetTitle()),
		}
	}

	// we may have updated the tool group config
	cfg := ToolGroupConfig{
		ID:           generated.Id,
		Title:        diffStringTemplate(generated.GetTitle(), updated.GetTitle()),
		DisplayOrder: diffInt(generated.GetDisplayOrder(), updated.GetDisplayOrder()),
		Tools:        extractToolGroupLinkConfigs(generated.Tools, updated.Tools),
		Visible:      diffBool(generated.GetVisible(), updated.GetVisible()),
	}

	if !cfg.hasChanges() {
		return nil
	}

	return &cfg
}

func extractToolGroupLinkConfigs(generated, updated []*toolsproto.ToolGroup_GroupActionLink) ToolGroupLinkConfigs {
	cfgs := ToolGroupLinkConfigs{}

	// we will use a map to keep track of the generated links that have already been configured
	availableLinks := map[string]bool{}
	for _, l := range generated {
		availableLinks[l.ActionLink.ToolId] = true
	}

	for _, l := range updated {
		if available, ok := availableLinks[l.ActionLink.ToolId]; ok && available {
			// if we have a generated link that hasn't been configured yet (still available), let's use that and diff it with our updated link
			if cfg := extractToolGroupLinkConfig(toolsproto.FindToolGroupLinkByToolID(generated, l.ActionLink.ToolId), l); cfg != nil {
				cfgs = append(cfgs, cfg)
			}
			// mark it as used
			availableLinks[l.ActionLink.ToolId] = false

			continue
		}

		// we don't have an available link, so we will add new ones
		cfgs = append(cfgs, extractToolGroupLinkConfig(nil, l))
	}

	// if there are any generated links that are still available, it means we've removed them, so let's add config for that as well
	for id, available := range availableLinks {
		if available {
			cfgs = append(cfgs, extractToolGroupLinkConfig(toolsproto.FindToolGroupLinkByToolID(generated, id), nil))
		}
	}

	if len(cfgs) > 0 {
		return cfgs
	}

	return nil
}

func extractToolGroupLinkConfig(generated, updated *toolsproto.ToolGroup_GroupActionLink) *ToolGroupLinkConfig {
	// we don't have a link and we didn't add a link
	if generated == nil && updated == nil {
		return nil
	}
	// we have a link and we've removed it
	if generated != nil && updated == nil {
		return &ToolGroupLinkConfig{
			ActionLink: &LinkConfig{ToolID: generated.ActionLink.ToolId},
			Deleted:    diffBool(false, true),
		}
	}

	// we didn't have a link, and now we've added it
	if generated == nil && updated != nil {
		return &ToolGroupLinkConfig{
			ActionLink:        extractLinkConfig(nil, updated.ActionLink),
			ResponseOverrides: updated.GetResponseOverridesMap(),
		}
	}

	linkDiff := extractLinkConfig(generated.ActionLink, updated.ActionLink)
	if linkDiff == nil {
		linkDiff = &LinkConfig{
			ToolID: updated.ActionLink.ToolId,
		}
	}

	cfg := ToolGroupLinkConfig{
		ActionLink:        linkDiff,
		ResponseOverrides: updated.GetResponseOverridesMap(),
	}

	if !cfg.hasChanges() {
		return nil
	}

	return &cfg
}

func diffString(old, new string) *string {
	if old != new {
		return &new
	}

	return nil
}

func diffInt(old, new int32) *int32 {
	if old != new {
		return &new
	}

	return nil
}

func diffBool(old, new bool) *bool {
	if old != new {
		return &new
	}

	return nil
}

func diffStringTemplate(old, new *toolsproto.StringTemplate) *string {
	if new != nil {
		if old != nil && old.Template == new.Template {
			return nil
		}
		return &new.Template
	}
	return nil
}
