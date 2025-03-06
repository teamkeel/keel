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
	for _, t := range v.Tools.Tools {
		v.validateTool(t)
	}
}

func (v *Validator) validateTool(t *toolsproto.ActionConfig) bool {
	hasError := false

	for _, l := range t.AllActionLinks() {
		if v.validateActionLink(l) {
			hasError = true
		}
	}

	if hasError {
		t.HasErrors = true
	}

	return hasError
}

func (v *Validator) validateActionLink(link *toolsproto.ActionLink) bool {
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
