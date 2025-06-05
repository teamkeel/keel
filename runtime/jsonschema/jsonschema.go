package jsonschema

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/xeipuuv/gojsonschema"
)

var (
	AnyTypes = []string{"string", "object", "array", "integer", "number", "boolean", "null"}
)

var (
	pageInfoSchema = JSONSchema{
		Properties: map[string]JSONSchema{
			"count": {
				Type: "number",
			},
			"startCursor": {
				Type: "string",
			},
			"endCursor": {
				Type: "string",
			},
			"totalCount": {
				Type: "number",
			},
			"hasNextPage": {
				Type: "boolean",
			},
			"pageNumber": {
				Type: "number",
			},
		},
	}
)

type JSONSchema struct {
	// Type is generally just a string, but when we need a type to be
	// null it is a list containing the type and the string "null".
	// In JSON output for most cases we just want a string, not a list
	// of one string, so we use any here so it can be either
	Type any `json:"type,omitempty"`

	// New in draft 6
	// Const is used to restrict a value to a single value.
	// https://json-schema.org/understanding-json-schema/reference/const
	Const any `json:"const,omitempty"`

	// The enum field needs to be able to contains strings and null,
	// so we use *string here
	Enum []*string `json:"enum,omitempty"`

	// Validation for strings
	Format string `json:"format,omitempty"`

	// Validation for objects
	Properties            map[string]JSONSchema `json:"properties,omitempty"`
	AdditionalProperties  *bool                 `json:"additionalProperties,omitempty"`
	UnevaluatedProperties *bool                 `json:"unevaluatedProperties,omitempty"`
	Required              []string              `json:"required,omitempty"`
	OneOf                 []JSONSchema          `json:"oneOf,omitempty"`
	AnyOf                 []JSONSchema          `json:"anyOf,omitempty"`
	Title                 string                `json:"title,omitempty"`
	Default               string                `json:"default,omitempty"`

	// For arrays
	Items *JSONSchema `json:"items,omitempty"`

	// Used to link to a type defined in the root $defs
	Ref string `json:"$ref,omitempty"`

	// Only used in the root JSONSchema object to define types that
	// can then be referenced using $ref
	Components  *Components           `json:"components,omitempty"`
	Definitions map[string]JSONSchema `json:"$defs,omitempty"`
}

type Components struct {
	Schemas map[string]JSONSchema `json:"schemas"`
}

// ValidateRequest validates that the input is valid for the given action and schema.
// If validation errors are found they will be contained in the returned result. If an error
// is returned then validation could not be completed, likely to do an invalid JSON schema
// being created.
func ValidateRequest(ctx context.Context, schema *proto.Schema, action *proto.Action, input any) (*gojsonschema.Result, error) {
	requestSchema := JSONSchema{}
	if action.InputMessageName != "" {
		requestSchema = JSONSchemaForActionInput(ctx, schema, action)
	}

	// We want to allow ISO8601 format WITH a compulsory date component to be permitted for the date format
	gojsonschema.FormatCheckers.Add("date", RelaxedDateFormatChecker{})

	return gojsonschema.Validate(gojsonschema.NewGoLoader(requestSchema), gojsonschema.NewGoLoader(input))
}

func ValidateResponse(ctx context.Context, schema *proto.Schema, action *proto.Action, response any) (JSONSchema, *gojsonschema.Result, error) {
	responseSchema := JSONSchemaForActionResponse(ctx, schema, action)
	result, err := gojsonschema.Validate(gojsonschema.NewGoLoader(responseSchema), gojsonschema.NewGoLoader(response))
	return responseSchema, result, err
}

func JSONSchemaForActionInput(ctx context.Context, schema *proto.Schema, action *proto.Action) JSONSchema {
	inputMessage := schema.FindMessage(action.GetInputMessageName())
	return JSONSchemaForMessage(ctx, schema, action, inputMessage, true)
}

func JSONSchemaForActionResponse(ctx context.Context, schema *proto.Schema, action *proto.Action) JSONSchema {
	if action.GetResponseMessageName() != "" {
		responseMsg := schema.FindMessage(action.GetResponseMessageName())

		return JSONSchemaForMessage(ctx, schema, action, responseMsg, false)
	}

	// If we've reached this point then we know that we are dealing with built-in actions
	switch action.GetType() {
	case proto.ActionType_ACTION_TYPE_CREATE, proto.ActionType_ACTION_TYPE_GET, proto.ActionType_ACTION_TYPE_UPDATE:
		// these action types return the serialized model

		model := schema.FindModel(action.GetModelName())

		if len(action.GetResponseEmbeds()) > 0 {
			return objectSchemaForModel(ctx, schema, model, false, action.GetResponseEmbeds())
		}

		return jsonSchemaForModel(ctx, schema, model, false)
	case proto.ActionType_ACTION_TYPE_LIST:
		// array of models

		model := schema.FindModel(action.GetModelName())

		modelSchema := JSONSchema{}
		if len(action.GetResponseEmbeds()) > 0 {
			modelSchema = objectSchemaForModel(ctx, schema, model, false, action.GetResponseEmbeds())
		} else {
			modelSchema = jsonSchemaForModel(ctx, schema, model, true)
		}

		// as there are nested components within the modelSchema, we need to merge these into the top level
		components := Components{
			Schemas: map[string]JSONSchema{},
		}
		for key, prop := range modelSchema.Components.Schemas {
			components.Schemas[key] = prop
		}
		modelSchema.Components = nil

		wrapperSchema := JSONSchema{
			Properties: map[string]JSONSchema{
				"results":  modelSchema,
				"pageInfo": pageInfoSchema,
			},
			Components: &components,
		}

		resultInfoSchema := jsonSchemaForFacets(schema, action)
		if resultInfoSchema != nil {
			wrapperSchema.Properties["resultInfo"] = *resultInfoSchema
		}

		return wrapperSchema
	case proto.ActionType_ACTION_TYPE_DELETE:
		// string id of deleted record

		return JSONSchema{
			Type: "string",
		}
	default:
		return JSONSchema{}
	}
}

func jsonSchemaForFacets(schema *proto.Schema, action *proto.Action) *JSONSchema {
	facetFields := proto.FacetFields(schema, action)
	if len(facetFields) == 0 {
		return nil
	}

	facetSchema := JSONSchema{
		Properties: map[string]JSONSchema{},
	}

	for _, field := range facetFields {
		switch field.GetType().GetType() {
		case proto.Type_TYPE_DECIMAL, proto.Type_TYPE_INT:
			facetSchema.Properties[field.GetName()] = JSONSchema{
				Properties: map[string]JSONSchema{
					"min": {
						Type: "number",
					},
					"max": {
						Type: "number",
					},
					"avg": {
						Type: "number",
					},
				},
			}
		case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
			facetSchema.Properties[field.GetName()] = JSONSchema{
				Properties: map[string]JSONSchema{
					"min": {
						Type:   "string",
						Format: "date-time",
					},
					"max": {
						Type:   "string",
						Format: "date-time",
					},
				},
			}
		case proto.Type_TYPE_DATE:
			facetSchema.Properties[field.GetName()] = JSONSchema{
				Properties: map[string]JSONSchema{
					"min": {
						Type:   "string",
						Format: "date",
					},
					"max": {
						Type:   "string",
						Format: "date",
					},
				},
			}
		case proto.Type_TYPE_DURATION:
			facetSchema.Properties[field.GetName()] = JSONSchema{
				Properties: map[string]JSONSchema{
					"min": {
						Type:   "string",
						Format: "duration",
					},
					"max": {
						Type:   "string",
						Format: "duration",
					},
				},
			}
		case proto.Type_TYPE_ENUM, proto.Type_TYPE_STRING:
			facetSchema.Properties[field.GetName()] = JSONSchema{
				Type: "array",
				Items: &JSONSchema{
					Type: "object",
					Properties: map[string]JSONSchema{
						"value": {
							Type: "string",
						},
						"count": {
							Type: "number",
						},
					},
				},
			}
		}
	}

	return &facetSchema
}

func contains(s []*proto.MessageField, e string) bool {
	for _, input := range s {
		if input.GetName() == e {
			return true
		}
	}

	return false
}

// Generates JSONSchema for an operation by generating properties for the root input message.
// Any subsequent nested messages are referenced.
func JSONSchemaForMessage(ctx context.Context, schema *proto.Schema, action *proto.Action, message *proto.Message, isInput bool) JSONSchema {
	components := Components{
		Schemas: map[string]JSONSchema{},
	}

	messageIsNil := message == nil
	isAny := !messageIsNil && message.GetName() == parser.MessageFieldTypeAny

	root := JSONSchema{
		Type:                  "object",
		Properties:            map[string]JSONSchema{},
		UnevaluatedProperties: boolPtr(false),
	}

	if isAny {
		anyOf := []JSONSchema{}
		for _, v := range AnyTypes {
			anyOf = append(anyOf, JSONSchema{Title: v, Type: v})
		}

		root.AnyOf = anyOf
		root.Type = nil
	}

	if !isAny {
		// Certain messages should only allow one field to be set per request so we set these as a oneOf property
		oneOfGroupFields := (len(message.GetFields()) == 3 && contains(message.GetFields(), "equals") && contains(message.GetFields(), "notEquals") && contains(message.GetFields(), "oneOf"))
		oneOfConditions := message.GetName() == "StringQueryInput" || message.GetName() == "BooleanQueryInput" || oneOfGroupFields
		// For these query inputs, we should allow multiple fields
		anyOfConditions := message.GetName() == "DateQueryInput" || message.GetName() == "TimestampQueryInput" || message.GetName() == "IntQueryInput"

		if oneOfConditions || anyOfConditions {
			jsonSchema := []JSONSchema{}

			for _, field := range message.GetFields() {
				jsonSchemaOption := JSONSchema{
					Type:       "object",
					Properties: map[string]JSONSchema{},
				}

				prop := jsonSchemaForField(ctx, schema, action, field.GetType(), field.GetNullable(), []string{}, isInput)

				// Merge components from this request schema into OpenAPI components
				if prop.Components != nil {
					for name, comp := range prop.Components.Schemas {
						components.Schemas[name] = comp
					}
					prop.Components = nil
				}

				jsonSchemaOption.Properties[field.GetName()] = prop
				jsonSchemaOption.Title = field.GetName()
				jsonSchemaOption.Required = append(jsonSchemaOption.Required, field.GetName())
				// https://json-schema.org/understanding-json-schema/reference/object#unevaluatedproperties

				jsonSchema = append(jsonSchema, jsonSchemaOption)
			}

			if anyOfConditions {
				root.AnyOf = jsonSchema
			} else {
				root.OneOf = jsonSchema
			}

			root.Type = nil
		} else {
			for _, field := range message.GetFields() {
				prop := jsonSchemaForField(ctx, schema, action, field.GetType(), field.GetNullable(), []string{}, isInput)

				// Merge components from this request schema into OpenAPI components
				if prop.Components != nil {
					for name, comp := range prop.Components.Schemas {
						components.Schemas[name] = comp
					}
					prop.Components = nil
				}

				root.Properties[field.GetName()] = prop

				// If the input is not optional then mark it required in the JSON schema
				if !field.GetOptional() {
					root.Required = append(root.Required, field.GetName())
				}
			}
		}
	}

	if len(components.Schemas) > 0 {
		root.Components = &components
	}

	return root
}

func jsonSchemaForModel(ctx context.Context, schema *proto.Schema, model *proto.Model, isRepeated bool) JSONSchema {
	definitionSchema := JSONSchema{
		Properties: map[string]JSONSchema{},
	}

	s := JSONSchema{}
	components := &Components{
		Schemas: map[string]JSONSchema{},
	}

	if isRepeated {
		s.Type = "array"
		s.Items = &JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", model.GetName())}
	} else {
		s = JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", model.GetName())}
	}

	for _, field := range model.GetFields() {
		// if the field of model type, then we don't want to include this because JSON-based
		// apis don't serialize nested relations
		if field.GetType().GetType() == proto.Type_TYPE_MODEL {
			continue
		}

		fieldSchema := jsonSchemaForField(ctx, schema, nil, field.GetType(), field.GetOptional(), []string{}, false)

		definitionSchema.Properties[field.GetName()] = fieldSchema

		// If the field is not optional then mark it as required in the JSON schema
		if !field.GetOptional() {
			definitionSchema.Required = append(definitionSchema.Required, field.GetName())
		}
	}

	schemas := map[string]JSONSchema{}

	components.Schemas[model.GetName()] = definitionSchema

	schemas[model.GetName()] = definitionSchema

	s.Components = components

	return s
}

func objectSchemaForModel(ctx context.Context, schema *proto.Schema, model *proto.Model, isRepeated bool, embeddings []string) JSONSchema {
	s := JSONSchema{
		Type:       "object",
		Properties: map[string]JSONSchema{},
	}
	components := &Components{
		Schemas: map[string]JSONSchema{},
	}

	for _, field := range model.GetFields() {
		fieldEmbeddings := []string{}

		// if the field is of ID type, and the related model is embedded, we do not want to include it in the schema
		if field.GetType().GetType() == proto.Type_TYPE_ID && field.GetForeignKeyInfo() != nil {
			relatedModel := strings.TrimSuffix(field.GetName(), "Id")
			skip := false
			for _, embed := range embeddings {
				frags := strings.Split(embed, ".")
				if frags[0] == relatedModel {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		// if the field of model type, and the model is not included in embeddings,
		// then we don't want to include this
		if field.GetType().GetType() == proto.Type_TYPE_MODEL {
			found := false

			for _, embed := range embeddings {
				frags := strings.Split(embed, ".")
				if frags[0] == field.GetName() {
					found = true
					// if we have to embed a child model for this field, we need to pass them through the field schema
					// with the first segment removed
					if len(frags) > 1 {
						fieldEmbeddings = append(fieldEmbeddings, strings.Join(frags[1:], "."))
					}
				}
			}

			if !found {
				continue
			}
		}

		fieldSchema := jsonSchemaForField(ctx, schema, nil, field.GetType(), field.GetOptional(), fieldEmbeddings, false)
		// If that nested field component has ref fields itself, then its components must be bundled.
		if fieldSchema.Components != nil {
			for cName, comp := range fieldSchema.Components.Schemas {
				components.Schemas[cName] = comp
			}
			fieldSchema.Components = nil
		}

		s.Properties[field.GetName()] = fieldSchema

		// If the field is not optional then mark it as required in the JSON schema
		if !field.GetOptional() {
			s.Required = append(s.Required, field.GetName())
		}
	}

	s.Components = components

	if isRepeated {
		return JSONSchema{
			Type:       "array",
			Items:      &s,
			Components: s.Components,
		}
	}
	return s
}

func jsonSchemaForField(ctx context.Context, schema *proto.Schema, action *proto.Action, t *proto.TypeInfo, isNullableField bool, embeddings []string, isInput bool) JSONSchema {
	components := &Components{
		Schemas: map[string]JSONSchema{},
	}
	prop := JSONSchema{}

	switch t.GetType() {
	case proto.Type_TYPE_ANY:
		anyOf := []JSONSchema{}
		for _, v := range AnyTypes {
			anyOf = append(anyOf, JSONSchema{Title: v, Type: v})
		}
		prop.AnyOf = anyOf
	case proto.Type_TYPE_MESSAGE:
		// Add the nested message to schema components.
		message := schema.FindMessage(t.GetMessageName().GetValue())
		component := JSONSchemaForMessage(ctx, schema, action, message, isInput)

		// If that nested message component has ref fields itself, then its components must be bundled.
		if component.Components != nil {
			for cName, comp := range component.Components.Schemas {
				components.Schemas[cName] = comp
			}
			component.Components = nil
		}

		name := t.GetMessageName().GetValue()
		if isNullableField {
			component.allowNull()
			name = "Nullable" + name
		}

		if t.GetRepeated() {
			prop.Type = "array"
			prop.Items = &JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", name)}
		} else {
			prop = JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", name)}
		}

		components.Schemas[name] = component

	case proto.Type_TYPE_UNION:
		// Union types can be modelled using oneOf.
		oneOf := []JSONSchema{}
		for _, m := range t.GetUnionNames() {
			// Add the nested message to schema components.
			message := schema.FindMessage(m.GetValue())
			component := JSONSchemaForMessage(ctx, schema, action, message, isInput)

			// Components of oneOf properties should only have one field per property and we should set a title.
			oneOfFieldName := message.GetFields()[0].GetName()
			component.Title = oneOfFieldName

			// If that nested message component has ref fields itself, then its components must be bundled.
			if component.Components != nil {
				for cName, comp := range component.Components.Schemas {
					components.Schemas[cName] = comp
				}
				component.Components = nil
			}

			name := message.GetName()
			if isNullableField {
				component.allowNull()
				name = "Nullable" + name
			}

			j := JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", name)}
			oneOf = append(oneOf, j)

			components.Schemas[name] = component
		}

		if t.GetRepeated() {
			prop.Type = "array"
			prop.Items = &JSONSchema{OneOf: oneOf}
		} else {
			prop = JSONSchema{OneOf: oneOf}
		}

	case proto.Type_TYPE_ID, proto.Type_TYPE_STRING:
		prop.Type = "string"
	case proto.Type_TYPE_MARKDOWN:
		prop.Type = "string"
		prop.Format = "markdown"
	case proto.Type_TYPE_DURATION:
		prop.Type = "string"
		prop.Format = "duration"
	case proto.Type_TYPE_BOOL:
		prop.Type = "boolean"
	case proto.Type_TYPE_INT:
		prop.Type = "number"
	case proto.Type_TYPE_DECIMAL:
		prop.Type = "number"
		prop.Format = "float"
	case proto.Type_TYPE_MODEL:
		model := schema.FindModel(t.GetModelName().GetValue())

		modelSchema := JSONSchema{}
		if len(embeddings) > 0 {
			modelSchema = objectSchemaForModel(ctx, schema, model, t.GetRepeated(), embeddings)
		} else {
			modelSchema = jsonSchemaForModel(ctx, schema, model, t.GetRepeated())
		}

		// If that nested message component has ref fields itself, then its components must be bundled.
		if modelSchema.Components != nil {
			for cName, comp := range modelSchema.Components.Schemas {
				components.Schemas[cName] = comp
			}
			modelSchema.Components = nil
		}

		if len(embeddings) > 0 {
			prop = modelSchema
		} else {
			if t.GetRepeated() {
				prop.Items = &JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", model.GetName())}
				prop.Type = "array"
			} else {
				prop = JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", model.GetName())}
			}
		}
	case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		// date-time format allows both YYYY-MM-DD and full ISO8601/RFC3339 format
		prop.Type = "string"
		prop.Format = "date-time"
	case proto.Type_TYPE_DATE:
		prop.Type = "string"
		prop.Format = "date"
	case proto.Type_TYPE_RELATIVE_PERIOD:
		prop.Type = "string"
	case proto.Type_TYPE_ENUM:
		// For enum's we actually don't need to set the `type` field at all
		enum, _ := lo.Find(schema.GetEnums(), func(e *proto.Enum) bool {
			return e.GetName() == t.GetEnumName().GetValue()
		})

		for _, v := range enum.GetValues() {
			prop.Enum = append(prop.Enum, &v.Name)
		}

		if isNullableField {
			prop.allowNull()
		}
	case proto.Type_TYPE_SORT_DIRECTION:
		prop.Type = "string"
		asc := "asc"
		desc := "desc"
		prop.Enum = []*string{&asc, &desc}
	case proto.Type_TYPE_FILE:
		// if the field is used as an input, so the type will be a data-url
		if isInput {
			prop.Type = "string"
			prop.Format = "data-url"
		} else {
			// if the field is as part of a response, then the action is nil and we want to return an object
			prop.Type = "object"
			prop.Properties = map[string]JSONSchema{
				"key":         {Type: "string"},
				"filename":    {Type: "string"},
				"contentType": {Type: "string"},
				"size":        {Type: "number"},
				"url":         {Type: "string"},
			}
			prop.Required = []string{"key", "filename", "contentType", "size", "url"}
		}
	}

	if t.GetRepeated() && (t.GetType() != proto.Type_TYPE_MESSAGE && t.GetType() != proto.Type_TYPE_MODEL && t.GetType() != proto.Type_TYPE_UNION) {
		prop.Items = &JSONSchema{Type: prop.Type, Enum: prop.Enum, Format: prop.Format}
		prop.Enum = nil
		prop.Format = ""
		prop.Type = "array"
	}

	if isNullableField {
		prop.allowNull()
	}

	if len(components.Schemas) > 0 {
		prop.Components = components
	}

	return prop
}

// allowNull makes sure that it allows null, either by modifying
// the type field or the enum field
//
// This is an area where OpenAPI differs from JSON Schema, from
// the OpenAPI spec:
//
//	| Note that there is no null type; instead, the nullable
//	| attribute is used as a modifier of the base type.
//
// We currently only support JSON schema.
func (s *JSONSchema) allowNull() {
	t := s.Type
	switch t := t.(type) {
	case string:
		s.Type = []string{t, "null"}
	case []string:
		if lo.Contains(t, "null") {
			return
		}
		t = append(t, "null")
		s.Type = t
	}

	if len(s.Enum) > 0 && !lo.Contains(s.Enum, nil) {
		s.Enum = append(s.Enum, nil)
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func ErrorsToString(errs []gojsonschema.ResultError) (ret string) {
	for _, err := range errs {
		ret += fmt.Sprintf("%s\n", err.String())
	}

	return ret
}

type RelaxedDateFormatChecker struct{}

// Checks that the value matches the a ISO8601 except the date component is mandatory.
func (f RelaxedDateFormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	formats := []string{
		"2006-01-02",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if _, err := time.Parse(format, asString); err == nil {
			return true
		}
	}

	return false
}
