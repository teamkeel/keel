package completions

import (
	"fmt"

	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

type PositionalContext struct {
	Completions []*CompletionItem `json:"completions"`
}

type CompletionItem struct {
	Description string `json:"description"`
	Label       string `json:"label"`
	Node        Node   `json:"node"`
}

type Node struct {
	Name   string        `json:"name"`
	Pos    node.Position `json:"pos"`
	EndPos node.Position `json:"end_pos"`
}

func ProvideCompletions(ast *parser.AST, position node.Position) (completions []*CompletionItem) {
	clearCompletions := func() {
		completions = make([]*CompletionItem, 0)
	}

	astInRange := ast.InRange(position)
	completions = append(completions, topLevelKeywordCompletions(ast)...)

	// If there is whitespace at the start/end of the file, this won't be included in the range of AST wrapper
	// so in that case, just return the default top level completion keywords
	if !astInRange {
		return completions
	}

	for _, declaration := range ast.Declarations {
		switch {
		case declaration.API != nil && declaration.API.InRange(position):
			clearCompletions()

			completions = append(completions, apiKeywordCompletions(declaration.API)...)

			for _, sect := range declaration.API.Sections {
				sectionInRange := sect.InRange(position)

				if sectionInRange {
					clearCompletions()
				}

				switch {
				case sect.Models != nil, sect.Node.Tokens[0].Value == parser.KeywordModels:
					// todo: exclude any existing models in models{} from the suggestions
					completions = append(completions, modelNamesForApiModelsCompletions(ast)...)
				}
			}
		case declaration.Model != nil && declaration.Model.InRange(position):
			clearCompletions()
			completions = append(completions, modelKeywordCompletions(declaration.Model)...)

			if declaration.Model != nil {
				for _, section := range declaration.Model.Sections {
					switch {
					case section.Fields != nil, section.InRange(position):
						clearCompletions()

						// todo: check for position of field/type. probably need to allow for incomplete field definitions first

						// inside of field blocks, provide attribute suggestions
						for _, field := range section.Fields {
							fieldInRange := field.InRange(position)

							if fieldInRange {
								clearCompletions()

								completions = append(completions, inFieldsBlockCompletions(field)...)
							}
						}
					case section.Functions != nil, section.Node.Tokens[0].Value == parser.KeywordFunctions:
						// todo
					case section.Operations != nil, section.Node.Tokens[0].Value == parser.KeywordOperations:
						// todo
					}
				}
			}

		case declaration.Role != nil && declaration.Role.InRange(position):
			clearCompletions()
			completions = append(completions, roleKeywordCompletions(declaration.Role)...)
		case declaration.Enum != nil && declaration.Enum.InRange(position):
			clearCompletions()

			// todo: dont think there are any completions to provide for enum values

		default:
			// if there is a syntax error within, it will fail to parse further
			// so we need to check the first token of the containing section

			firstToken := declaration.Tokens[0].Value

			switch firstToken {
			case parser.KeywordApi:
				completions = append(completions, apiKeywordCompletions(declaration.API)...)

			case parser.KeywordModel:
				completions = append(completions, modelKeywordCompletions(declaration.Model)...)
			case parser.KeywordRole:
				completions = append(completions, roleKeywordCompletions(declaration.Role)...)
			}
		}
	}

	return completions
}

func modelNamesForApiModelsCompletions(ast *parser.AST) (completions []*CompletionItem) {
	for _, model := range query.Models([]*parser.AST{ast}) {
		completions = append(completions, &CompletionItem{
			Label: model.Name.Value,
		})
	}

	return completions
}

func inFieldsBlockCompletions(node parser.GenericNode) (completions []*CompletionItem) {
	keywords := []string{fmt.Sprintf("@%s", parser.AttributeUnique), fmt.Sprintf("@%s", parser.AttributeDefault)}
	return stringArrayToCompletionsArray(keywords, node)
}

func roleKeywordCompletions(node parser.GenericNode) (completions []*CompletionItem) {
	keywords := []string{parser.KeywordEmails, parser.KeywordDomains}
	return stringArrayToCompletionsArray(keywords, node)
}

func modelKeywordCompletions(node parser.GenericNode) (completions []*CompletionItem) {
	keywords := []string{
		parser.KeywordFields,
		parser.KeywordFunctions,
		parser.KeywordOperations,
		fmt.Sprintf("@%s", parser.AttributePermission),
	}
	return stringArrayToCompletionsArray(keywords, node)
}

func apiKeywordCompletions(node parser.GenericNode) (completions []*CompletionItem) {
	keywords := []string{parser.KeywordModels, fmt.Sprintf("@%s", parser.AttributeGraphQL)}
	return stringArrayToCompletionsArray(keywords, node)
}

func topLevelKeywordCompletions(node parser.GenericNode) (completions []*CompletionItem) {
	keywords := []string{
		parser.KeywordModel,
		parser.KeywordApi,
		parser.KeywordRole,
		parser.KeywordEnum,
	}

	for _, keyword := range keywords {

		completions = append(completions, &CompletionItem{
			Label: keyword,
			Node:  buildNode(node),
		})
	}

	return completions
}

func keywordMatchingFirstToken(keyword string, node parser.GenericNode) bool {
	tokens := node.GetTokens()

	if len(tokens) >= 1 {
		if tokens[0].Value == keyword {
			return true
		}
	}

	return false
}

func stringArrayToCompletionsArray(arr []string, node parser.GenericNode) (completions []*CompletionItem) {
	for _, item := range arr {
		completions = append(completions, &CompletionItem{
			Label:       item,
			Node:        buildNode(node),
			Description: fmt.Sprintf("Available in %s", node.String()),
		})
	}

	return completions
}

func buildNode(n parser.GenericNode) Node {
	start, end := n.GetPositionRange()

	return Node{
		Name: n.String(),
		Pos: node.Position{
			Line:   start.Line,
			Column: start.Column,
		},
		EndPos: node.Position{
			Line:   end.Line,
			Column: end.Column,
		},
	}
}
