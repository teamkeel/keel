package tools

import (
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

func extractConfig(generated, updated *toolsproto.Tool) *ToolConfig {
	if generated.GetType() != updated.GetType() {
		return nil
	}

	var cfg *ToolConfig

	if generated.IsActionBased() {
		cfg = extractActionConfig(generated.GetActionConfig(), updated.GetActionConfig()).toToolConfig()
	} else {
		cfg = extractFlowConfig(generated.GetFlowConfig(), updated.GetFlowConfig()).toToolConfig()
	}
	cfg.setID(updated.GetId())
	return cfg
}

func extractFlowConfig(generated, updated *toolsproto.FlowConfig) *FlowToolConfig {
	cfg := &FlowToolConfig{
		FlowName:           updated.GetFlowName(),
		Name:               diffString(generated.GetName(), updated.GetName()),
		HelpText:           diffStringTemplate(generated.GetHelpText(), updated.GetHelpText()),
		CompletionRedirect: extractLinkConfig(generated.GetCompletionRedirect(), updated.GetCompletionRedirect()),
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
		ID:                   updated.GetId(),
		ActionName:           updated.GetActionName(),
		Name:                 diffString(generated.GetName(), updated.GetName()),
		Icon:                 diffString(generated.GetIcon(), updated.GetIcon()),
		Title:                diffStringTemplate(generated.GetTitle(), updated.GetTitle()),
		HelpText:             diffStringTemplate(generated.GetHelpText(), updated.GetHelpText()),
		EntitySingle:         diffString(generated.GetEntitySingle(), updated.GetEntitySingle()),
		EntityPlural:         diffString(generated.GetEntityPlural(), updated.GetEntityPlural()),
		Pagination:           extractPaginationConfig(generated.GetPagination(), updated.GetPagination()),
		CreateEntryAction:    extractLinkConfig(generated.GetCreateEntryAction(), updated.GetCreateEntryAction()),
		GetEntryAction:       extractLinkConfig(generated.GetGetEntryAction(), updated.GetGetEntryAction()),
		EntryActivityActions: extractLinkConfigs(generated.GetEntryActivityActions(), updated.GetEntryActivityActions()),
		RelatedActions:       extractLinkConfigs(generated.GetRelatedActions(), updated.GetRelatedActions()),
		EmbeddedTools:        extractToolGroupConfigs(generated.GetEmbeddedTools(), updated.GetEmbeddedTools()),
		FilterConfig:         extractFilterConfig(generated.GetFilterConfig(), updated.GetFilterConfig()),
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
			Label:            el.GetLabel().GetTemplate(),
			Href:             el.GetHref().GetTemplate(),
			Icon:             el.Icon,
			DisplayOrder:     el.GetDisplayOrder(),
			VisibleCondition: el.VisibleCondition,
		})
	}
	cfg.Sections = Sections{}
	for _, s := range updated.GetSections() {
		cfg.Sections = append(cfg.Sections, &Section{
			Name:             s.GetName(),
			Title:            s.GetTitle().GetTemplate(),
			Description:      diffStringTemplate(nil, s.GetDescription()),
			VisibleCondition: s.VisibleCondition,
			DisplayOrder:     s.GetDisplayOrder(),
			Visible:          s.GetVisible(),
		})
	}

	if generated.GetDisplayLayout().JSON() != updated.GetDisplayLayout().JSON() {
		cfg.DisplayLayout = &DisplayLayoutConfig{
			Config: updated.GetDisplayLayout().AsObj(),
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
		LookupAction:     extractLinkConfig(generated.GetLookupAction(), updated.GetLookupAction()),
		GetEntryAction:   extractLinkConfig(generated.GetGetEntryAction(), updated.GetGetEntryAction()),
	}

	if updated.GetDefaultValue() != nil {
		switch updated.GetDefaultValue().GetValue().(type) {
		case *toolsproto.ScalarValue_Bool:
			val := updated.GetDefaultValue().GetBool()
			cfg.DefaultValue = &ScalarValue{BoolValue: &val}
		case *toolsproto.ScalarValue_Float:
			val := updated.GetDefaultValue().GetFloat()
			cfg.DefaultValue = &ScalarValue{FloatValue: &val}
		case *toolsproto.ScalarValue_String_:
			val := updated.GetDefaultValue().GetString_()
			cfg.DefaultValue = &ScalarValue{StringValue: &val}
		case *toolsproto.ScalarValue_Integer:
			val := updated.GetDefaultValue().GetInteger()
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

	if updated.GetDefaultValue() != nil {
		switch updated.GetDefaultValue().GetValue().(type) {
		case *toolsproto.ScalarValue_Bool:
			val := updated.GetDefaultValue().GetBool()
			cfg.DefaultValue = &ScalarValue{BoolValue: &val}
		case *toolsproto.ScalarValue_Float:
			val := updated.GetDefaultValue().GetFloat()
			cfg.DefaultValue = &ScalarValue{FloatValue: &val}
		case *toolsproto.ScalarValue_String_:
			val := updated.GetDefaultValue().GetString_()
			cfg.DefaultValue = &ScalarValue{StringValue: &val}
		case *toolsproto.ScalarValue_Integer:
			val := updated.GetDefaultValue().GetInteger()
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
		Link:             extractLinkConfig(generated.GetLink(), updated.GetLink()),
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
		availableLinks[l.GetToolId()] = true
	}

	for _, l := range updated {
		if available, ok := availableLinks[l.GetToolId()]; ok && available {
			// if we have a generated link that hasn't been configured yet (still available), let's use that and diff it with our updated link
			if cfg := extractLinkConfig(toolsproto.FindLinkByToolID(generated, l.GetToolId()), l); cfg != nil {
				cfgs = append(cfgs, cfg)
			}
			// mark it as used
			availableLinks[l.GetToolId()] = false

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

	if generated != nil && updated != nil {
		return nil
	}

	if generated == nil && updated != nil {
		if updated.GetPageSize() == nil {
			return nil
		}

		return &PaginationConfig{
			PageSize: extractPageSizeConfig(nil, updated.GetPageSize()),
		}
	}

	pageSize := extractPageSizeConfig(generated.GetPageSize(), updated.GetPageSize())

	if pageSize == nil {
		return nil
	}

	return &PaginationConfig{
		PageSize: extractPageSizeConfig(generated.GetPageSize(), updated.GetPageSize()),
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
			DefaultValue: updated.DefaultValue,
		}
	}

	defaultValue := diffNullableInt(generated.DefaultValue, updated.DefaultValue)

	if defaultValue == nil {
		return nil
	}

	return &PageSizeConfig{
		DefaultValue: defaultValue,
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
			ToolID:  generated.GetToolId(),
			Deleted: diffBool(false, true),
		}
	}

	// we didn't have a link, and now we've added it
	if generated == nil && updated != nil {
		return &LinkConfig{
			ToolID:           updated.GetToolId(),
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
		ToolID:           generated.GetToolId(),
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
			if cfg := extractToolGroupConfig(toolsproto.FindToolGroupByID(generated, g.GetId()), g); cfg != nil {
				cfgs = append(cfgs, cfg)
			}
			// mark it as used
			availableGroups[g.GetId()] = false

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
			ID:           updated.GetId(),
			Visible:      &updated.Visible,
			Tools:        extractToolGroupLinkConfigs(nil, updated.GetTools()),
			DisplayOrder: &updated.DisplayOrder,
			Title:        diffStringTemplate(nil, updated.GetTitle()),
		}
	}

	// we may have updated the tool group config
	cfg := ToolGroupConfig{
		ID:           generated.GetId(),
		Title:        diffStringTemplate(generated.GetTitle(), updated.GetTitle()),
		DisplayOrder: diffInt(generated.GetDisplayOrder(), updated.GetDisplayOrder()),
		Tools:        extractToolGroupLinkConfigs(generated.GetTools(), updated.GetTools()),
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
		availableLinks[l.GetActionLink().GetToolId()] = true
	}

	for _, l := range updated {
		if available, ok := availableLinks[l.GetActionLink().GetToolId()]; ok && available {
			// if we have a generated link that hasn't been configured yet (still available), let's use that and diff it with our updated link
			if cfg := extractToolGroupLinkConfig(toolsproto.FindToolGroupLinkByToolID(generated, l.GetActionLink().GetToolId()), l); cfg != nil {
				cfgs = append(cfgs, cfg)
			}
			// mark it as used
			availableLinks[l.GetActionLink().GetToolId()] = false

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
			ActionLink: &LinkConfig{ToolID: generated.GetActionLink().GetToolId()},
			Deleted:    diffBool(false, true),
		}
	}

	// we didn't have a link, and now we've added it
	if generated == nil && updated != nil {
		return &ToolGroupLinkConfig{
			ActionLink:        extractLinkConfig(nil, updated.GetActionLink()),
			ResponseOverrides: updated.GetResponseOverridesMap(),
		}
	}

	linkDiff := extractLinkConfig(generated.GetActionLink(), updated.GetActionLink())
	if linkDiff == nil {
		linkDiff = &LinkConfig{
			ToolID: updated.GetActionLink().GetToolId(),
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

func diffString(old, updated string) *string {
	if old != updated {
		return &updated
	}

	return nil
}

func diffInt(old, updated int32) *int32 {
	if old != updated {
		return &updated
	}

	return nil
}

func diffNullableInt(old, updated *int32) *int32 {
	if old != nil && updated != nil && *old == *updated {
		return nil
	}

	if old != nil && updated == nil {
		return updated
	}

	return updated
}

func diffBool(old, updated bool) *bool {
	if old != updated {
		return &updated
	}

	return nil
}

func diffStringTemplate(old, updated *toolsproto.StringTemplate) *string {
	if updated != nil {
		if old != nil && old.GetTemplate() == updated.GetTemplate() {
			return nil
		}
		return &updated.Template
	}
	return nil
}
