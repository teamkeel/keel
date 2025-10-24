package mcpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/jsonschema"
)

// generateResourceDescription creates a human-readable description for a read action resource
func generateResourceDescription(action *proto.Action, model *proto.Model) string {
	switch action.GetType() {
	case proto.ActionType_ACTION_TYPE_GET:
		return fmt.Sprintf("Get a single %s record by %s", model.Name, getGetActionLookupDescription(action))
	case proto.ActionType_ACTION_TYPE_LIST:
		filters := getListActionFilters(action)
		if len(filters) > 0 {
			return fmt.Sprintf("List %s records filtered by: %s", model.Name, strings.Join(filters, ", "))
		}
		return fmt.Sprintf("List all %s records with pagination support", model.Name)
	case proto.ActionType_ACTION_TYPE_READ:
		return fmt.Sprintf("Read %s data with custom query logic", model.Name)
	default:
		return fmt.Sprintf("Read %s data", model.Name)
	}
}

// generateToolDescription creates a human-readable description for a write action tool
func generateToolDescription(action *proto.Action, model *proto.Model, schema *proto.Schema) string {
	switch action.GetType() {
	case proto.ActionType_ACTION_TYPE_CREATE:
		fields := getCreateActionFields(action, model, schema)
		if len(fields) > 0 {
			return fmt.Sprintf("Create a new %s with: %s", model.Name, strings.Join(fields, ", "))
		}
		return fmt.Sprintf("Create a new %s record", model.Name)
	case proto.ActionType_ACTION_TYPE_UPDATE:
		fields := getUpdateActionFields(action, model, schema)
		if len(fields) > 0 {
			return fmt.Sprintf("Update a %s by %s. Can update: %s", model.Name, getGetActionLookupDescription(action), strings.Join(fields, ", "))
		}
		return fmt.Sprintf("Update a %s record", model.Name)
	case proto.ActionType_ACTION_TYPE_DELETE:
		return fmt.Sprintf("Delete a %s by %s", model.Name, getGetActionLookupDescription(action))
	case proto.ActionType_ACTION_TYPE_WRITE:
		return fmt.Sprintf("Perform a write operation on %s data with custom logic", model.Name)
	default:
		return fmt.Sprintf("Modify %s data", model.Name)
	}
}

// getGetActionLookupDescription describes what fields are used to look up a record
func getGetActionLookupDescription(action *proto.Action) string {
	if action.InputMessageName == "" {
		return "ID"
	}

	// Get input message to understand lookup fields
	inputFields := getInputFieldNames(action)
	if len(inputFields) == 0 {
		return "ID"
	}

	if len(inputFields) == 1 {
		return inputFields[0]
	}

	return strings.Join(inputFields, " and ")
}

// getListActionFilters describes available filters for a list action
func getListActionFilters(action *proto.Action) []string {
	if action.InputMessageName == "" {
		return nil
	}

	inputFields := getInputFieldNames(action)
	return inputFields
}

// getCreateActionFields returns the fields that can be set in a create action
func getCreateActionFields(action *proto.Action, model *proto.Model, schema *proto.Schema) []string {
	if action.InputMessageName == "" {
		// If no input message, list writable fields from model
		return getWritableFieldNames(model)
	}

	inputFields := getInputFieldNames(action)
	return inputFields
}

// getUpdateActionFields returns the fields that can be updated
func getUpdateActionFields(action *proto.Action, model *proto.Model, schema *proto.Schema) []string {
	if action.InputMessageName == "" {
		return getWritableFieldNames(model)
	}

	// Get input message fields, excluding lookup fields
	inputFields := getInputFieldNames(action)

	// Filter out ID-like fields that are typically for lookup
	updateFields := []string{}
	for _, field := range inputFields {
		if !isLookupField(field) {
			updateFields = append(updateFields, field)
		}
	}

	return updateFields
}

// getInputFieldNames returns field names from an action's input message
func getInputFieldNames(action *proto.Action) []string {
	// For implicit inputs, we'd need to look at the message definition
	// For now, we can return a basic implementation
	// In a full implementation, we'd parse the input message from schema
	return []string{}
}

// getWritableFieldNames returns non-system field names from a model
func getWritableFieldNames(model *proto.Model) []string {
	fields := []string{}
	for _, field := range model.Fields {
		// Skip system fields
		if isSystemField(field.Name) {
			continue
		}
		fields = append(fields, field.Name)
	}
	return fields
}

// isSystemField checks if a field is a system-managed field
func isSystemField(fieldName string) bool {
	systemFields := map[string]bool{
		"id":        true,
		"createdAt": true,
		"updatedAt": true,
	}
	return systemFields[fieldName]
}

// isLookupField checks if a field name suggests it's used for lookups
func isLookupField(fieldName string) bool {
	lookupFields := map[string]bool{
		"id": true,
	}
	return lookupFields[fieldName] || strings.HasSuffix(fieldName, "Id")
}

// generateInputSchema creates a JSON Schema for an action's inputs using the jsonschema package
func generateInputSchema(action *proto.Action, model *proto.Model, schema *proto.Schema) map[string]interface{} {
	// Use the existing jsonschema package to generate a proper JSON schema
	ctx := context.Background()
	inputSchema := jsonschema.JSONSchemaForActionInput(ctx, schema, action)

	// Convert JSONSchema struct to map[string]interface{} for MCP protocol
	schemaBytes, err := json.Marshal(inputSchema)
	if err != nil {
		// Fallback to minimal schema if marshaling fails
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &result); err != nil {
		// Fallback to minimal schema if unmarshaling fails
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	// MCP requires the root type to always be the string literal "object", not an array
	// The jsonschema package can return ["object", "null"] for nullable types
	if typeVal, ok := result["type"]; ok {
		switch t := typeVal.(type) {
		case []interface{}:
			// If type is an array, extract the non-null type
			for _, item := range t {
				if str, ok := item.(string); ok && str != "null" {
					result["type"] = str
					break
				}
			}
		}
	}

	// Ensure type is always set to "object" at root level
	if result["type"] == nil || result["type"] == "" {
		result["type"] = "object"
	}

	// Inline all $ref references - MCP expects self-contained schemas
	// Extract components before inlining
	var components map[string]interface{}
	if comps, ok := result["components"]; ok {
		if compsMap, ok := comps.(map[string]interface{}); ok {
			if schemas, ok := compsMap["schemas"]; ok {
				if schemasMap, ok := schemas.(map[string]interface{}); ok {
					components = schemasMap
				}
			}
		}
	}

	// Inline all references recursively
	inlineRefs(result, components)

	// Now remove components since everything is inlined
	delete(result, "components")
	delete(result, "$defs")

	// Claude API doesn't support oneOf/allOf/anyOf at the top level
	// If we have these at root, wrap them in a proper object schema
	if hasTopLevelComposite(result) {
		result = wrapCompositeSchema(result)
	}

	// Remove unsupported JSON Schema features for Claude API compatibility
	removeUnsupportedFeatures(result)

	return result
}

// inlineRefs recursively resolves $ref fields by replacing them with the actual schema
func inlineRefs(obj map[string]interface{}, components map[string]interface{}) {
	for key, value := range obj {
		if key == "$ref" {
			// This is a reference - resolve it
			if refStr, ok := value.(string); ok {
				// Parse reference like "#/components/schemas/CreateInput"
				if strings.HasPrefix(refStr, "#/components/schemas/") {
					schemaName := strings.TrimPrefix(refStr, "#/components/schemas/")
					if components != nil {
						if refSchema, ok := components[schemaName]; ok {
							// Replace the parent object with the referenced schema
							// Copy all fields from refSchema into obj
							if refMap, ok := refSchema.(map[string]interface{}); ok {
								// Clear current object and copy all fields from referenced schema
								for k := range obj {
									delete(obj, k)
								}
								for k, v := range refMap {
									obj[k] = deepCopy(v)
								}
								// Continue inlining in the newly copied schema
								inlineRefs(obj, components)
								return
							}
						}
					}
				}
			}
		}

		// Recurse into nested structures
		switch v := value.(type) {
		case map[string]interface{}:
			inlineRefs(v, components)
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					inlineRefs(m, components)
				}
			}
		}
	}
}

// deepCopy creates a deep copy of a value
func deepCopy(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		copy := make(map[string]interface{})
		for k, v := range val {
			copy[k] = deepCopy(v)
		}
		return copy
	case []interface{}:
		copy := make([]interface{}, len(val))
		for i, item := range val {
			copy[i] = deepCopy(item)
		}
		return copy
	default:
		return v
	}
}

// removeUnsupportedFeatures recursively removes JSON Schema features not supported by Claude API
func removeUnsupportedFeatures(schema map[string]interface{}) {
	// Remove unsupported keywords at this level
	delete(schema, "unevaluatedProperties")
	delete(schema, "$schema")
	delete(schema, "$id")
	delete(schema, "examples")

	// Recursively process nested objects
	for _, value := range schema {
		switch v := value.(type) {
		case map[string]interface{}:
			removeUnsupportedFeatures(v)
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					removeUnsupportedFeatures(m)
				}
			}
		}
	}
}

// hasTopLevelComposite checks if the schema has oneOf/allOf/anyOf at the top level
func hasTopLevelComposite(schema map[string]interface{}) bool {
	_, hasOneOf := schema["oneOf"]
	_, hasAllOf := schema["allOf"]
	_, hasAnyOf := schema["anyOf"]
	return hasOneOf || hasAllOf || hasAnyOf
}

// wrapCompositeSchema wraps a schema with oneOf/allOf/anyOf at root into a proper object
func wrapCompositeSchema(schema map[string]interface{}) map[string]interface{} {
	// Create a new object schema that contains the composite as a property
	// This way Claude can still understand and use the schema
	wrapped := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}

	// If there's a oneOf/allOf/anyOf, try to merge them intelligently
	if oneOf, hasOneOf := schema["oneOf"]; hasOneOf {
		// For oneOf at root, we'll try to extract common properties
		// or just flatten to the first option if it's simple
		if options, ok := oneOf.([]interface{}); ok && len(options) > 0 {
			// Try to merge all properties from all options
			properties := make(map[string]interface{})
			required := []string{}

			for _, opt := range options {
				if optMap, ok := opt.(map[string]interface{}); ok {
					if props, ok := optMap["properties"].(map[string]interface{}); ok {
						for k, v := range props {
							properties[k] = v
						}
					}
					// Don't enforce required fields since they vary across oneOf options
				}
			}

			if len(properties) > 0 {
				wrapped["properties"] = properties
				if len(required) > 0 {
					wrapped["required"] = required
				}
			}
		}
		delete(schema, "oneOf")
	}

	if allOf, hasAllOf := schema["allOf"]; hasAllOf {
		// For allOf, merge all schemas together
		if options, ok := allOf.([]interface{}); ok {
			properties := make(map[string]interface{})
			requiredSet := make(map[string]bool)

			for _, opt := range options {
				if optMap, ok := opt.(map[string]interface{}); ok {
					if props, ok := optMap["properties"].(map[string]interface{}); ok {
						for k, v := range props {
							properties[k] = v
						}
					}
					if req, ok := optMap["required"].([]interface{}); ok {
						for _, r := range req {
							if str, ok := r.(string); ok {
								requiredSet[str] = true
							}
						}
					}
				}
			}

			if len(properties) > 0 {
				wrapped["properties"] = properties
				if len(requiredSet) > 0 {
					required := []string{}
					for k := range requiredSet {
						required = append(required, k)
					}
					wrapped["required"] = required
				}
			}
		}
		delete(schema, "allOf")
	}

	if anyOf, hasAnyOf := schema["anyOf"]; hasAnyOf {
		// For anyOf, similar to oneOf - merge properties
		if options, ok := anyOf.([]interface{}); ok {
			properties := make(map[string]interface{})

			for _, opt := range options {
				if optMap, ok := opt.(map[string]interface{}); ok {
					if props, ok := optMap["properties"].(map[string]interface{}); ok {
						for k, v := range props {
							properties[k] = v
						}
					}
				}
			}

			if len(properties) > 0 {
				wrapped["properties"] = properties
			}
		}
		delete(schema, "anyOf")
	}

	// Copy over any remaining properties from the original schema
	for k, v := range schema {
		if k != "oneOf" && k != "allOf" && k != "anyOf" && k != "type" {
			wrapped[k] = v
		}
	}

	return wrapped
}

// generateOutputDescription creates a description of what an action returns
func generateOutputDescription(action *proto.Action, model *proto.Model) string {
	switch action.GetType() {
	case proto.ActionType_ACTION_TYPE_GET:
		return fmt.Sprintf("Returns a single %s record", model.Name)
	case proto.ActionType_ACTION_TYPE_LIST:
		return fmt.Sprintf("Returns a paginated list of %s records", model.Name)
	case proto.ActionType_ACTION_TYPE_CREATE:
		return fmt.Sprintf("Returns the newly created %s record", model.Name)
	case proto.ActionType_ACTION_TYPE_UPDATE:
		return fmt.Sprintf("Returns the updated %s record", model.Name)
	case proto.ActionType_ACTION_TYPE_DELETE:
		return fmt.Sprintf("Returns the ID of the deleted %s", model.Name)
	case proto.ActionType_ACTION_TYPE_READ:
		return fmt.Sprintf("Returns custom %s data", model.Name)
	case proto.ActionType_ACTION_TYPE_WRITE:
		return fmt.Sprintf("Returns the result of the write operation", )
	default:
		return "Returns the result of the operation"
	}
}
