package attribute

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// attributeLocationsRule checks that attributes are used in valid places
// For example it's invalid to use a @where attribute inside a model definition.
func AttributeLocationsRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, section := range model.Sections {
			if section.Attribute != nil {
				errs.Concat(checkAttributes([]*parser.AttributeNode{section.Attribute}, "model", model.Name.Value))
			}

			if section.Actions != nil {
				for _, function := range section.Actions {
					errs.Concat(checkAttributes(function.Attributes, parser.KeywordActions, function.Name.Value))
				}
			}

			if section.Fields != nil {
				for _, field := range section.Fields {
					errs.Concat(checkAttributes(field.Attributes, "field", field.Name.Value))
				}
			}
		}
	}

	for _, task := range query.Tasks(asts) {
		for _, section := range task.Sections {
			if section.Attribute != nil {
				errs.Concat(checkAttributes([]*parser.AttributeNode{section.Attribute}, "task", task.Name.Value))
			}

			if section.Fields != nil {
				for _, field := range section.Fields {
					errs.Concat(checkAttributes(field.Attributes, "field", field.Name.Value))
				}
			}
		}
	}

	for _, job := range query.Jobs(asts) {
		for _, section := range job.Sections {
			if section.Attribute != nil {
				errs.Concat(checkAttributes([]*parser.AttributeNode{section.Attribute}, "job", job.Name.Value))
			}
		}
	}

	for _, api := range query.APIs(asts) {
		for _, section := range api.Sections {
			if section.Attribute != nil {
				errs.Concat(checkAttributes([]*parser.AttributeNode{section.Attribute}, "api", api.Name.Value))
			}
		}
	}

	for _, flow := range query.Flows(asts) {
		for _, section := range flow.Sections {
			if section.Attribute != nil {
				errs.Concat(checkAttributes([]*parser.AttributeNode{section.Attribute}, "flow", flow.Name.Value))
			}
		}
	}

	return
}

var attributeLocations = map[string][]string{
	parser.KeywordModel: {
		parser.AttributePermission,
		parser.AttributeUnique,
		parser.AttributeOn,
	},
	parser.KeywordField: {
		parser.AttributeUnique,
		parser.AttributeDefault,
		parser.AttributePrimaryKey,
		parser.AttributeRelation,
		parser.AttributeComputed,
		parser.AttributeSequence,
	},
	parser.KeywordActions: {
		parser.AttributeSet,
		parser.AttributeWhere,
		parser.AttributePermission,
		parser.AttributeValidate,
		parser.AttributeOrderBy,
		parser.AttributeSortable,
		parser.AttributeFunction,
		parser.AttributeEmbed,
		parser.AttributeFacet,
	},
	parser.KeywordJob: {
		parser.AttributePermission,
		parser.AttributeSchedule,
	},
	parser.KeywordFlow: {
		parser.AttributePermission,
		parser.AttributeSchedule,
	},
}

func checkAttributes(attributes []*parser.AttributeNode, definedOn string, parentName string) (errs errorhandling.ValidationErrors) {
	for _, attr := range attributes {
		allowedAttributes := attributeLocations[definedOn]

		if lo.Contains(allowedAttributes, attr.Name.Value) {
			continue
		}

		hintOptions := []string{}

		for _, allowed := range allowedAttributes {
			hintOptions = append(hintOptions, fmt.Sprintf("@%s", allowed))
		}

		hint := errorhandling.NewCorrectionHint(hintOptions, attr.Name.Value)
		suggestions := formatting.HumanizeList(hint.Results, formatting.DelimiterOr)

		errs.Append(errorhandling.ErrorUnsupportedAttributeType,
			map[string]string{
				"Name":        attr.Name.Value,
				"ParentName":  parentName,
				"DefinedOn":   definedOn,
				"Suggestions": suggestions,
			},
			attr.Name,
		)
	}

	return
}
