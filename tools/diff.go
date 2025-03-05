package tools

import toolsproto "github.com/teamkeel/keel/tools/proto"

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
