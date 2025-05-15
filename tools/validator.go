package tools

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

type Validator struct {
	Schema *proto.Schema
	Tools  *toolsproto.Tools
}

// NewValidator creates a new tool validator for the given schema and tools
func NewValidator(schema *proto.Schema, tools *toolsproto.Tools) *Validator {
	return &Validator{
		Schema: schema,
		Tools:  tools,
	}
}

func (v *Validator) validate() {
	for _, t := range v.Tools.Configs {
		if t.Type == toolsproto.Tool_ACTION {
			v.validateActionConfig(t.ActionConfig)
		}
	}
}

func (v *Validator) validateActionConfig(t *toolsproto.ActionConfig) bool {
	hasError := false
	// first let's validate all top level action links
	toolLinks := []*toolsproto.ToolLink{}
	if t.CreateEntryAction != nil {
		toolLinks = append(toolLinks, t.CreateEntryAction)
	}
	if t.GetEntryAction != nil {
		toolLinks = append(toolLinks, t.GetEntryAction)
	}
	toolLinks = append(toolLinks, t.RelatedActions...)
	toolLinks = append(toolLinks, t.EntryActivityActions...)
	for _, l := range toolLinks {
		hasError = hasError || v.validateToolLink(l)
	}

	// now we validate inputs & response fields
	for _, in := range t.Inputs {
		hasError = hasError || v.validateInput(in)
	}
	for _, out := range t.Response {
		hasError = hasError || v.validateResponse(out)
	}

	// now we validate tool groups
	for _, tg := range t.GetEmbeddedTools() {
		hasError = hasError || v.validateToolGroup(tg)
	}

	// now we validate display layouts
	if dl := t.GetDisplayLayout(); dl != nil {
		hasError = hasError || v.validateDisplayLayout(dl)
	}

	// // validate that the underlying action exists
	if v.Schema.FindAction(t.ActionName) == nil {
		t.Errors = append(t.Errors, &toolsproto.ValidationError{
			Error: fmt.Sprintf("Data source does not exist: %s", t.ActionName),
			Field: "action_name",
		})
		hasError = true
	}

	if hasError {
		t.HasErrors = true
	}

	return hasError
}

// validateDisplayLayout will validate the given display layout and if applicable, it will add a Validation error to it.
// Returns true if an error has been detected.
func (v *Validator) validateDisplayLayout(dl *toolsproto.DisplayLayoutConfig) bool {
	hasError := false
	for _, link := range dl.AllToolLinks() {
		hasError = hasError || v.validateToolLink(link)
	}

	dl.HasErrors = hasError

	return hasError
}

// validateToolGroup will validate the given group and if applicable, it will add a Validation error to it.
// Returns true if an error has been detected.
func (v *Validator) validateToolGroup(tg *toolsproto.ToolGroup) bool {
	hasError := false
	for _, tgl := range tg.GetTools() {
		if tgl.GetActionLink() != nil {
			hasError = hasError || v.validateToolLink(tgl.GetActionLink())
		}
	}

	tg.HasErrors = hasError

	return hasError
}

// validateInput will validate the given input and if applicable, it will add a Validation error to it.
// Returns true if an error has been detected.
func (v *Validator) validateInput(input *toolsproto.RequestFieldConfig) bool {
	hasError := false
	if input.LookupAction != nil {
		hasError = hasError || v.validateToolLink(input.LookupAction)
	}
	if input.GetEntryAction != nil {
		hasError = hasError || v.validateToolLink(input.GetEntryAction)
	}
	input.HasErrors = hasError

	return hasError
}

// validateResponse will validate the given response field and if applicable, it will add a Validation error to it.
// Returns true if an error has been detected.
func (v *Validator) validateResponse(out *toolsproto.ResponseFieldConfig) bool {
	hasError := false
	if out.Link != nil {
		hasError = hasError || v.validateToolLink(out.Link)
	}
	out.HasErrors = hasError

	return hasError
}

// validateToolLink will validate the given action link and if applicable, it will add a Validation error to it.
// Returns true if an error has been detected.
func (v *Validator) validateToolLink(link *toolsproto.ToolLink) bool {
	hasError := false
	if link == nil {
		return false
	}

	// validate that the target tool exists
	if targetTool := v.Tools.FindByID(link.ToolId); targetTool == nil {
		// target tool doesn't exist
		hasError = true
		link.Errors = append(link.Errors, &toolsproto.ValidationError{
			Error: fmt.Sprintf("Target tool does not exist: %s", link.ToolId),
			Field: "tool_id",
		})
	}

	return hasError
}
