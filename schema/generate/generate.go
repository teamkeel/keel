package generate

import (
	"fmt"
	"strings"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func GenerateDefaultActions(asts []*parser.AST, modelName string) []*parser.ActionNode {
	model := query.Model(asts, modelName)
	if model == nil {
		return []*parser.ActionNode{}
	}

	fields := query.ModelFields(model, query.ExcludeBuiltInFields)

	lowerModelName := casing.ToLowerCamel(modelName)
	pluralModelName := casing.ToLowerCamel(casing.ToPlural(modelName))

	actions := []*parser.ActionNode{}

	/**
	Action generation logic:

	 - Get action that takes id

	 - List action that takes all fields as inputs
	 	- All inputs are optional
	 	- 1:m relationships included as filters

	 - Create action that takes all fields as inputs
	 	- Nullable fields are optional, non-nullable fields are required
	 	- Ignore 1:m relationship fields

	 - Update action that takes all fields
	 	- All inputs are optional
	 	- Ignore 1:m relationship fields

	 - Delete action that takes id
	**/

	actions = append(actions, genGetAction(lowerModelName))
	actions = append(actions, genListAction(asts, pluralModelName, fields))
	actions = append(actions, genCreateAction(asts, lowerModelName, fields))
	actions = append(actions, genUpdateAction(asts, lowerModelName, fields))
	actions = append(actions, genDeleteAction(lowerModelName))

	return actions
}

func GenerateDefaultActionsAsStrings(asts []*parser.AST, modelName string) []string {
	actions := GenerateDefaultActions(asts, modelName)

	result := make([]string, len(actions))
	for i, action := range actions {
		result[i] = formatAction(action)
	}

	return result
}

func genGetAction(modelName string) *parser.ActionNode {
	return &parser.ActionNode{
		Type: parser.NameNode{Value: parser.ActionTypeGet},
		Name: parser.NameNode{Value: "get" + casing.ToCamel(modelName)},
		Inputs: []*parser.ActionInputNode{
			genActionInput("id", false),
		},
	}
}

func genListAction(asts []*parser.AST, pluralModelName string, fields []*parser.FieldNode) *parser.ActionNode {
	var inputs []*parser.ActionInputNode

	for _, field := range fields {
		if field.Repeated {
			// For repeated fields (HasMany relationships), add relationshipField.id
			if query.IsModel(asts, field.Type.Value) {
				inputs = append(inputs, genRelationshipInput(field.Name.Value, true))
			}
			continue
		} else {
			// For non-repeated fields
			input := genFieldInput(asts, field, true) // All list inputs are optional
			if input != nil {
				inputs = append(inputs, input)
			}
		}
	}

	return &parser.ActionNode{
		Type:   parser.NameNode{Value: parser.ActionTypeList},
		Name:   parser.NameNode{Value: "list" + casing.ToCamel(pluralModelName)},
		Inputs: inputs,
	}
}

func genCreateAction(asts []*parser.AST, modelName string, fields []*parser.FieldNode) *parser.ActionNode {
	var withInputs []*parser.ActionInputNode

	for _, field := range fields {
		if field.Repeated {
			continue // Skip repeated fields for create
		}

		input := genFieldInput(asts, field, field.Optional)
		if input != nil {
			withInputs = append(withInputs, input)
		}
	}

	return &parser.ActionNode{
		Type:   parser.NameNode{Value: parser.ActionTypeCreate},
		Name:   parser.NameNode{Value: "create" + casing.ToCamel(modelName)},
		Inputs: []*parser.ActionInputNode{}, // create actions typically have no inputs, just with clause
		With:   withInputs,
	}
}

func genUpdateAction(asts []*parser.AST, modelName string, fields []*parser.FieldNode) *parser.ActionNode {
	var withInputs []*parser.ActionInputNode

	for _, field := range fields {
		if field.Repeated {
			continue // Skip repeated fields for update
		}

		input := genFieldInput(asts, field, true) // All update inputs are optional
		if input != nil {
			withInputs = append(withInputs, input)
		}
	}

	return &parser.ActionNode{
		Type: parser.NameNode{Value: parser.ActionTypeUpdate},
		Name: parser.NameNode{Value: "update" + casing.ToCamel(modelName)},
		Inputs: []*parser.ActionInputNode{
			genActionInput("id", false),
		},
		With: withInputs,
	}
}

func genDeleteAction(modelName string) *parser.ActionNode {
	return &parser.ActionNode{
		Type: parser.NameNode{Value: parser.ActionTypeDelete},
		Name: parser.NameNode{Value: "delete" + casing.ToCamel(modelName)},
		Inputs: []*parser.ActionInputNode{
			genActionInput("id", false),
		},
	}
}

func genFieldInput(asts []*parser.AST, field *parser.FieldNode, optional bool) *parser.ActionInputNode {
	if field.IsScalar() {
		return genActionInput(field.Name.Value, optional)
	}
	if query.IsEnum(asts, field.Type.Value) {
		return genActionInput(field.Name.Value, optional)
	}
	if query.IsModel(asts, field.Type.Value) {
		return genRelationshipInput(field.Name.Value, optional)
	}
	return nil
}

func genActionInput(fieldName string, optional bool) *parser.ActionInputNode {
	return &parser.ActionInputNode{
		Type: parser.Ident{
			Fragments: []*parser.IdentFragment{
				{Fragment: fieldName},
			},
		},
		Optional: optional,
	}
}

func genRelationshipInput(fieldName string, optional bool) *parser.ActionInputNode {
	return &parser.ActionInputNode{
		Type: parser.Ident{
			Fragments: []*parser.IdentFragment{
				{Fragment: fieldName},
				{Fragment: "id"},
			},
		},
		Optional: optional,
	}
}

func formatAction(action *parser.ActionNode) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s %s", lowerCamel(action.Type.Value), lowerCamel(action.Name.Value)))

	formatActionInputs(&b, action.Inputs, false)

	if len(action.With) > 0 {
		b.WriteString(" with ")
		formatActionInputs(&b, action.With, false)
	}

	if len(action.Returns) > 0 {
		b.WriteString(" returns ")
		formatActionInputs(&b, action.Returns, false)
	}

	return strings.TrimSpace(b.String())
}

// TODO we might be able to leverage the format package here as there is a bit of duplication
func formatActionInputs(b *strings.Builder, inputs []*parser.ActionInputNode, isArbitraryFunction bool) {
	b.WriteString("(")

	if len(inputs) == 0 {
		b.WriteString(")")
		return
	}

	// Determine if we should use multiline format
	isMultiline := shouldUseMultilineFormat(inputs)

	if isMultiline {
		b.WriteString("\n")
		for i, input := range inputs {
			b.WriteString("    ") // 4 spaces for indentation

			// Format the input
			if input.Label != nil {
				b.WriteString(fmt.Sprintf("%s: %s", input.Label.Value, input.Type.Fragments[0].Fragment))
			} else {
				for j, fragment := range input.Type.Fragments {
					if j > 0 {
						b.WriteString(".")
					}
					if isArbitraryFunction {
						b.WriteString(fragment.Fragment)
					} else {
						b.WriteString(lowerCamel(fragment.Fragment))
					}
				}
			}

			if input.Optional {
				b.WriteString("?")
			}

			// Add comma if not the last input
			if i < len(inputs)-1 {
				b.WriteString(",")
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	} else {
		for i, input := range inputs {
			// Format the input
			if input.Label != nil {
				b.WriteString(fmt.Sprintf("%s: %s", input.Label.Value, input.Type.Fragments[0].Fragment))
			} else {
				for j, fragment := range input.Type.Fragments {
					if j > 0 {
						b.WriteString(".")
					}
					if isArbitraryFunction {
						b.WriteString(fragment.Fragment)
					} else {
						b.WriteString(lowerCamel(fragment.Fragment))
					}
				}
			}

			if input.Optional {
				b.WriteString("?")
			}

			// Add comma if not the last input
			if i < len(inputs)-1 {
				b.WriteString(", ")
			}
		}
	}

	b.WriteString(")")
}

func shouldUseMultilineFormat(inputs []*parser.ActionInputNode) bool {
	if len(inputs) <= 2 {
		return false
	}

	// Calculate total length to determine if we should use multiline
	totalLength := 2 // for parentheses
	for i, input := range inputs {
		if i > 0 {
			totalLength += 2 // for ", "
		}
		for j, fragment := range input.Type.Fragments {
			if j > 0 {
				totalLength += 1 // for "."
			}
			totalLength += len(fragment.Fragment)
		}
		if input.Optional {
			totalLength += 1 // for "?"
		}
	}

	return totalLength > 40
}

func lowerCamel(s string) string {
	return casing.ToLowerCamel(s)
}
