package tools

import (
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

func syncTool(tool, gen *toolsproto.ActionConfig) {
	if tool == nil || gen == nil {
		return
	}

	// we sync inputs
	newInputs := tool.DiffInputs(gen)
	for _, i := range tool.GetInputs() {
		if generated := gen.HasInput(i.GetFieldLocation()); generated != nil {
			syncInput(i, generated)
			newInputs = append(newInputs, i)
		}
	}
	tool.Inputs = newInputs

	// now we sync responses
	newResponses := tool.DiffResponse(gen)
	for _, i := range tool.GetResponse() {
		if generated := gen.HasResponse(i.GetFieldLocation()); generated != nil {
			syncResponse(i, generated)
			newResponses = append(newResponses, i)
		}
	}
	tool.Response = newResponses

	syncToolLinks(tool, gen)
}

func syncInput(input, gen *toolsproto.RequestFieldConfig) {
	input.FieldType = gen.FieldType
	input.Repeated = gen.Repeated
	input.EnumName = gen.EnumName
	input.ModelName = gen.ModelName
	input.FieldName = gen.FieldName
	input.Scope = gen.Scope

	// if we now have action links but we didn't before, let's bring them through
	if input.LookupAction == nil && gen.LookupAction != nil {
		input.LookupAction = gen.LookupAction
	}
	if input.GetEntryAction == nil && gen.GetEntryAction != nil {
		input.GetEntryAction = gen.GetEntryAction
	}
}

func syncResponse(resp, gen *toolsproto.ResponseFieldConfig) {
	resp.FieldType = gen.FieldType
	resp.Sortable = gen.Sortable
	resp.Repeated = gen.Repeated
	resp.EnumName = gen.EnumName
	resp.ModelName = gen.ModelName
	resp.FieldName = gen.FieldName
	resp.Scope = gen.Scope

	// if we now have action links but we didn't before, let's bring them through
	if resp.Link == nil && gen.Link != nil {
		resp.Link = gen.Link
	}
}

func syncToolLinks(tool, gen *toolsproto.ActionConfig) {
	// we add any new links that are in the generated tool but not in the original one
	if tool.CreateEntryAction == nil && gen.CreateEntryAction != nil {
		tool.CreateEntryAction = gen.CreateEntryAction
	}
	if tool.GetEntryAction == nil && gen.GetEntryAction != nil {
		tool.GetEntryAction = gen.GetEntryAction
	}

	for _, generatedLink := range gen.GetRelatedActions() {
		found := false
		for _, existing := range tool.GetRelatedActions() {
			if existing.ToolId == generatedLink.ToolId {
				found = true
				break
			}
		}

		if !found {
			tool.RelatedActions = append(tool.RelatedActions, generatedLink)
		}
	}

	for _, generatedLink := range gen.GetEntryActivityActions() {
		found := false
		for _, existing := range tool.GetEntryActivityActions() {
			if existing.ToolId == generatedLink.ToolId {
				found = true
				break
			}
		}

		if !found {
			tool.EntryActivityActions = append(tool.EntryActivityActions, generatedLink)
		}
	}

	// we add any tool groups that have been newly generated
	for _, generated := range gen.GetEmbeddedTools() {
		found := false
		for _, existing := range tool.GetEmbeddedTools() {
			if existing.Id == generated.Id {
				found = true
				break
			}
		}
		if !found {
			tool.EmbeddedTools = append(tool.EmbeddedTools, generated)
		}
	}
}
