package completions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/reader"
)

type CompletionItem struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Kind        string `json:"kind"`

	// If empty use `Label`
	InsertText string `json:"insertText"`
}

const (
	KindModel       = "model"
	KindField       = "field"
	KindVariable    = "variable"
	KindType        = "type"
	KindKeyword     = "keyword"
	KindLabel       = "label"
	KindAttribute   = "attribute"
	KindPunctuation = "punctuation"
	KindInput       = "inputs"
)

const (
	DescriptionSuggested = "Suggested"
)

func Completions(schemaFiles []*reader.SchemaFile, pos *node.Position, cfg *config.ProjectConfig) []*CompletionItem {

	var schema string
	asts := []*parser.AST{}

	for _, f := range schemaFiles {
		// parse the schema ignoring any errors, it's very likely the
		// schema is not in a valid state
		ast, _ := parser.Parse(f)
		asts = append(asts, ast)
		if f.FileName == pos.Filename {
			schema = f.Contents
		}
	}

	tokenAtPos := NewTokensAtPosition(schema, pos)

	// First check if we're within an attribute's arguments list
	// Attributes can appear in a number of places so easier to check
	// for this up-front
	_, isAttr := getParentAttribute(tokenAtPos)
	if isAttr {
		return getAttributeArgCompletions(asts, tokenAtPos, cfg)
	}

	enclosingBlock := getTypeOfEnclosingBlock(tokenAtPos)

	// switch on nearest (previous) keyword
	switch enclosingBlock {
	case parser.KeywordModel:
		attributes := getAttributeCompletions(tokenAtPos, []string{parser.AttributePermission, parser.AttributeUnique})
		return append(attributes, modelBlockKeywords...)
	case parser.KeywordRole:
		return roleBlockKeywords
	case parser.KeywordApi:
		return append([]*CompletionItem{}, apiBlockKeywords...)
	case parser.KeywordEnum:
		// no completions for enum block
		return []*CompletionItem{}
	case parser.KeywordFields:
		return getFieldCompletions(asts, tokenAtPos)
	case parser.KeywordMessage:
		return getMessageFieldCompletions(asts, tokenAtPos)
	case parser.KeywordActions:
		return getActionCompletions(asts, tokenAtPos, enclosingBlock)
	case parser.KeywordModels:
		// models block inside an api block - complete with model names
		return getUserDefinedTypeCompletions(asts, tokenAtPos, parser.KeywordModel)
	case parser.KeywordJob:
		attributes := getAttributeCompletions(tokenAtPos, []string{parser.AttributePermission, parser.AttributeSchedule})
		return append(attributes, getJobCompletions()...)
	case parser.KeywordInput:
		return getInputCompletions(asts, tokenAtPos)
	default:
		// If no enclosing block then we're at the top-level of the schema, or we are defining
		// a top level named block

		// if the previous token is one of the top level keywords then the current token
		// is a name and we can't provide any completions for that

		lastToken := tokenAtPos.ValueAt(-1)

		_, ok := lo.Find(topLevelKeywords, func(v *CompletionItem) bool {
			return v.Label == lastToken
		})

		if ok {
			if lastToken == parser.KeywordModel || lastToken == parser.KeywordEnum {
				return getUndefinedFieldCompletions(asts, tokenAtPos)
			} else {
				// api / role etc - no name completions possible
				return []*CompletionItem{}
			}
		}

		return topLevelKeywords
	}
}

func getUndefinedFieldCompletions(asts []*parser.AST, tokenAtPos *TokensAtPosition) (items []*CompletionItem) {
	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {
			// check that model exists
			model := query.Model(asts, field.Type.Value)

			enum := query.Enum(asts, field.Type.Value)

			if model == nil && enum == nil {
				items = append(items, &CompletionItem{
					Label: field.Type.Value,
					Kind:  KindType,
				})
			}
		}
	}

	return items
}

func getMessageFieldCompletions(asts []*parser.AST, tokenAtPos *TokensAtPosition) []*CompletionItem {
	// First we find the start of the current block
	startOfBlock := tokenAtPos.StartOfBlock()

	// Now we have to work out if we're expecting a field name, a field type
	curr := startOfBlock

	// First we expecting a field name
	expectingName := true

	for {
		// Move to the next token
		curr = curr.Next()

		// If this token is the token at the cursor position we can stop walking the tokens
		if curr.Is(tokenAtPos) {
			break
		}

		// If we were expecting a name then we're now expecting a type
		// so we flip the flag and move on
		if expectingName {
			expectingName = false
			continue
		}

		// otherwise we're now expecting a name again
		expectingName = true

		// skip past "?" token
		if curr.ValueAt(1) == "?" {
			curr = curr.Next()
		}

		// skip past "[]" tokens
		if curr.ValueAt(1) == "[" {
			curr = curr.Next().Next()
		}

		// skip past a field block
		if curr.ValueAt(1) == "{" {
			curr = curr.Next().EndOfBlock()
			continue
		}

	}

	// if we're expecting a name then we can't offer completions
	if expectingName {
		return []*CompletionItem{}
	}

	// Provide completions for field type which is built-in and user-defined types
	return lo.Flatten(
		[][]*CompletionItem{
			getUserDefinedTypeCompletions(asts, tokenAtPos, parser.KeywordModel),
			getUserDefinedTypeCompletions(asts, tokenAtPos, parser.KeywordEnum),
			getUserDefinedTypeCompletions(asts, tokenAtPos, parser.KeywordMessage),
			getUserDefinedTypeCompletions(asts, tokenAtPos, parser.KeywordJob),
			getBuiltInTypeCompletions(),
		},
	)
}

func getFieldCompletions(asts []*parser.AST, tokenAtPos *TokensAtPosition) []*CompletionItem {
	return getBlockCompletions(asts, tokenAtPos, parser.KeywordFields)
}

func getInputCompletions(asts []*parser.AST, tokenAtPos *TokensAtPosition) []*CompletionItem {
	return getBlockCompletions(asts, tokenAtPos, parser.KeywordInput)
}

func getBlockCompletions(asts []*parser.AST, tokenAtPos *TokensAtPosition, keyword string) []*CompletionItem {
	// First we find the start of the current block
	startOfBlock := tokenAtPos.StartOfBlock()

	// Simple case for field attributes:
	//   1. Current token is "@"
	//   2. Previous token is "@"
	//   3. Parent block is field block which can only contain attributes
	if tokenAtPos.Value() == "@" ||
		tokenAtPos.ValueAt(-1) == "@" ||
		startOfBlock.Prev().Value() != keyword {
		return getAttributeCompletions(tokenAtPos, []string{
			parser.AttributeUnique,
			parser.AttributeDefault,
			parser.AttributeRelation,
		})
	}

	// Now we have to work out if we're expecting a field name, a field type, or an
	// inline attribute (not enclosed in a block)
	curr := startOfBlock

	// First we expecting a field name
	expectingName := true

	for {
		// Move to the next token
		curr = curr.Next()

		// If this token is the token at the cursor position we can stop walking the tokens
		if curr.Is(tokenAtPos) {
			break
		}

		// If we were expecting a name then we're now expecting a type
		// so we flip the flag and move on
		if expectingName {
			expectingName = false
			continue
		}

		// otherwise we're now expecting a name again
		expectingName = true

		// skip past "?" token
		if curr.ValueAt(1) == "?" {
			curr = curr.Next()
		}

		// skip past "[]" tokens
		if curr.ValueAt(1) == "[" {
			curr = curr.Next().Next()
		}

		// skip past a field block
		if curr.ValueAt(1) == "{" {
			curr = curr.Next().EndOfBlock()
			continue
		}

		// skip past any attributes
	skippingAttributes:
		for {
			if curr.ValueAt(1) == "@" {
				curr = curr.Next() // go to "@"
				curr = curr.Next() // go to attribute name
				if curr.ValueAt(1) == "(" {
					curr = curr.Next().EndOfParen()
				}
				continue skippingAttributes
			}
			break skippingAttributes
		}
	}

	// if we're expecting a name then there are two cases
	if expectingName {
		// The current token is on the same line as the previous token
		// In this case we provide attribute name completions
		if tokenAtPos.Line() == tokenAtPos.Prev().Line() {
			return getAttributeCompletions(tokenAtPos, []string{
				parser.AttributeUnique,
				parser.AttributeDefault,
				parser.AttributeRelation,
			})
		}

		// We on a new line which means current token is field name for
		// which we can't provide completions
		return []*CompletionItem{}
	}

	results := []*CompletionItem{}
	results = append(results, getUserDefinedTypeCompletions(asts, tokenAtPos, parser.KeywordModel)...)
	results = append(results, getUserDefinedTypeCompletions(asts, tokenAtPos, parser.KeywordEnum)...)
	results = append(results, getBuiltInTypeCompletions()...)

	return results
}

func getActionCompletions(asts []*parser.AST, tokenAtPos *TokensAtPosition, enclosingBlock string) []*CompletionItem {
	// if we are inside enclosing parenthesis then we are completing for
	// action inputs, or for returns
	if tokenAtPos.StartOfParen() != nil {
		return getActionInputCompletions(asts, tokenAtPos)
	}

	// try to find the first matching token out of returns/read/write
	// prev is the first match that was found when searching backwards through the token stream
	// this will help identify the position (whether we are in a returns or in the inputs)
	prev, _ := tokenAtPos.FindPrevMultiple(
		parser.KeywordReturns,
		parser.ActionTypeRead,
		parser.ActionTypeWrite,
	)

	if tokenAtPos.Prev().EndOfParen() != nil && (prev == parser.ActionTypeWrite || prev == parser.ActionTypeRead) {
		// target just after end of inputs

		return []*CompletionItem{
			{
				Label: parser.KeywordReturns,
				Kind:  KindKeyword,
			},
		}
	}

	prev, _ = tokenAtPos.FindPrevMultipleOnLine(
		parser.ActionTypeDelete,
		parser.ActionTypeGet,
		parser.ActionTypeList,
		parser.KeywordWith,
	)
	// if we're delete, list or get action type and have completed our parenthesis, or there is already a `with`
	// clause on this line then there are no further completions that are valid. Return empty list.
	if tokenAtPos.Prev().EndOfParen() != nil && prev != "" {
		return []*CompletionItem{}
	}

	// to see if we should provide attribute completions we see if we're inside an
	// individual actions block or if the current or previous token is the "@"
	inActionBlock := tokenAtPos.ValueAt(-1) == "{" && tokenAtPos.ValueAt(-2) == ")"
	inAttribute := tokenAtPos.Value() == "@" || tokenAtPos.ValueAt(-1) == "@"

	if inActionBlock || inAttribute {
		return getAttributeCompletions(tokenAtPos, []string{
			parser.AttributeSet,
			parser.AttributeWhere,
			parser.AttributePermission,
			parser.AttributeValidate,
			parser.AttributeOrderBy,
			parser.AttributeSortable,
		})
	}

	// current token is action name - can't auto-complete
	if lo.Contains(parser.ActionTypes, tokenAtPos.ValueAt(-1)) {
		modelName := getParentModelName(tokenAtPos)
		actionType := tokenAtPos.ValueAt(-1)

		suggestion := fmt.Sprintf("%s%s", actionType, casing.ToCamel(modelName))
		return []*CompletionItem{{
			Label:       suggestion,
			Kind:        KindLabel,
			Description: DescriptionSuggested,
		}}
	}

	isNewLine := tokenAtPos.IsNewLine()
	isEndOfParen := tokenAtPos.Prev().EndOfParen() != nil

	return lo.Filter(actionBlockKeywords, func(i *CompletionItem, _ int) bool {
		// if we're at the start of a new line then we should remove `with`
		// if we're not on a new line and the last token is a closing parenthesis and we haven't got another
		// `with` on this line then the only valid completion is `with`
		if isNewLine {
			return i.Label != parser.KeywordWith
		} else if isEndOfParen {
			return i.Label == parser.KeywordWith
		}
		return true
	})
}

func getJobCompletions() []*CompletionItem {
	completions := []*CompletionItem{
		{
			Label:       "inputs",
			Description: "Inputs for manual runs",
			Kind:        KindInput,
		},
	}

	return completions
}

var builtInFieldCompletions = []*CompletionItem{
	{
		Label:       parser.ImplicitFieldNameId,
		Description: parser.FieldTypeID,
		Kind:        KindField,
	},
	{
		Label:       parser.ImplicitFieldNameCreatedAt,
		Description: parser.FieldTypeDatetime,
		Kind:        KindField,
	},
	{
		Label:       parser.ImplicitFieldNameUpdatedAt,
		Description: parser.FieldTypeDatetime,
		Kind:        KindField,
	},
}

var modelBlockKeywords = []*CompletionItem{
	{
		Label: parser.KeywordFields,
		Kind:  KindKeyword,
	},
	{
		Label: parser.KeywordActions,
		Kind:  KindKeyword,
	},
}

var actionBlockKeywords = []*CompletionItem{
	{
		Label: parser.ActionTypeCreate,
		Kind:  KindKeyword,
	},
	{
		Label: parser.ActionTypeUpdate,
		Kind:  KindKeyword,
	},
	{
		Label: parser.ActionTypeGet,
		Kind:  KindKeyword,
	},
	{
		Label: parser.ActionTypeList,
		Kind:  KindKeyword,
	},
	{
		Label: parser.ActionTypeDelete,
		Kind:  KindKeyword,
	},
	{
		Label: parser.KeywordWith,
		Kind:  KindKeyword,
	},
	{
		Label: parser.ActionTypeRead,
		Kind:  KindKeyword,
	},
	{
		Label: parser.ActionTypeWrite,
		Kind:  KindKeyword,
	},
}

var roleBlockKeywords = []*CompletionItem{
	{
		Label: parser.KeywordDomains,
		Kind:  KindKeyword,
	},
	{
		Label: parser.KeywordEmails,
		Kind:  KindKeyword,
	},
}

var apiBlockKeywords = []*CompletionItem{
	{
		Label: parser.KeywordModels,
		Kind:  KindKeyword,
	},
}

var topLevelKeywords = []*CompletionItem{
	{
		Label: parser.KeywordModel,
		Kind:  KindKeyword,
	},
	{
		Label: parser.KeywordRole,
		Kind:  KindKeyword,
	},
	{
		Label: parser.KeywordEnum,
		Kind:  KindKeyword,
	},
	{
		Label: parser.KeywordApi,
		Kind:  KindKeyword,
	},
	{
		Label: parser.KeywordMessage,
		Kind:  KindKeyword,
	},
	{
		Label: parser.KeywordJob,
		Kind:  KindKeyword,
	},
}

func getBuiltInTypeCompletions() []*CompletionItem {
	completions := []*CompletionItem{}
	for t := range parser.BuiltInTypes {
		completions = append(completions, &CompletionItem{
			Label:       t,
			Description: t,
			Kind:        KindType,
		})
	}
	completions = append(completions, &CompletionItem{
		Label: "Identity",
		Kind:  KindModel,
	})
	return completions
}

func getActionInputCompletions(asts []*parser.AST, tokenAtPos *TokensAtPosition) []*CompletionItem {
	// inside action input args - auto-complete field names
	completions := append([]*CompletionItem{}, builtInFieldCompletions...)

	block := tokenAtPos.StartOfBlock()

	actionType := block.Next().Value()

	readOrWriteActionType := actionType == parser.ActionTypeRead || actionType == parser.ActionTypeWrite

	if readOrWriteActionType {
		// find the first occurrence of a given token in the previous token stream and return what match was found
		match, _ := tokenAtPos.StartOfParen().FindPrevMultiple(
			parser.KeywordReturns,
			parser.ActionTypeRead,
			parser.ActionTypeWrite,
		)
		// if insideInputs is true, it means the first preceding token we're interested in was action type read or write
		insideInputs := match == parser.ActionTypeRead || match == parser.ActionTypeWrite

		// the first occurrence was the 'returns' keyword, so we know we're inside a returns parentheses
		insideReturns := match == parser.KeywordReturns

		messages := lo.Map(query.MessageNames(asts), func(msgName string, _ int) *CompletionItem {
			return &CompletionItem{
				Label: msgName,
				Kind:  KindLabel,
			}
		})

		switch {
		case insideInputs:
			// for read and write actions, we want to suggest the field names (default behaviour) and the available message types
			completions = append(completions, messages...)

			// in the proto generation, we append the 'Any' type as a new message to the proto schema's Messages
			// However we haven't reached the proto generation yet, so we are still in schema land
			// so it will be easiest just to append 'Any' here
			completions = append(completions, &CompletionItem{
				Label: "Any",
				Kind:  KindLabel,
			})
		case insideReturns:
			// suggest the message types available in the schema only

			return messages
		}
	}

	modelName := getParentModelName(tokenAtPos)
	model := query.Model(asts, modelName)

	// if we have been able to get the model from the AST we can try to
	// find the field names
	if model != nil {
		fieldNames := getModelFieldCompletions(model)

		// if current or previous token is a "." then we need to provide
		// completions for a related model
		if tokenAtPos.Value() == "." || tokenAtPos.ValueAt(-1) == "." {
			idents := getPreviousIdents(tokenAtPos)
			var ok bool
			fieldNames, ok = getFieldNamesAtPath(asts, model, idents)
			if !ok {
				// if we were unable to resolve the relevant model
				// return no completions as returning the default
				// fields in this case could be unhelpful
				return []*CompletionItem{}
			}
		}

		completions = append(completions, fieldNames...)
	}

	// if using long-form input syntax then current token can also be a built-in
	// type so add those
	if tokenAtPos.ValueAt(-1) == ":" {
		completions = append(completions, getBuiltInTypeCompletions()...)
	}

	return completions

}

func getAttributeArgCompletions(asts []*parser.AST, t *TokensAtPosition, cfg *config.ProjectConfig) []*CompletionItem {
	attrName, _ := getParentAttribute(t)

	enclosingBlock := getTypeOfEnclosingBlock(t)

	switch attrName {
	case parser.AttributeSet, parser.AttributeWhere, parser.AttributeValidate:
		return getExpressionCompletions(asts, t, cfg)
	case parser.AttributePermission:
		return getPermissionArgCompletions(asts, t, cfg)
	case parser.AttributeSortable:
		return getSortableArgCompletions(asts, t, cfg)
	case parser.AttributeOrderBy:
		return getOrderByArgCompletions(asts, t, cfg)
	case parser.AttributeSchedule:
		return getScheduleArgCompletions(asts, t, cfg)

	case parser.AttributeUnique:
		// composite
		if enclosingBlock == parser.KeywordModel {
			if t.Prev().Value() == parser.AttributeUnique {
				// open array notation

				return []*CompletionItem{{Label: "[", Kind: KindPunctuation}}
			}
			modelName := getParentModelName(t)
			model := query.Model(asts, modelName)

			fields := query.ModelFields(model, func(f *parser.FieldNode) bool {
				return f.IsScalar()
			})

			allFields := lo.Map(fields, func(f *parser.FieldNode, _ int) *CompletionItem {
				return &CompletionItem{
					Label: f.Name.Value,
					Kind:  KindVariable,
				}
			})

			// when the current token is an opening array bracket, start of list
			if t.Value() == "[" {
				return allFields
			}

			// when there are already items
			if t.Value() == "," {
				prev := t.Prev().Value()

				return lo.Filter(allFields, func(i *CompletionItem, _ int) bool {
					return i.Label != prev
				})
			}

			// if the previous token is one of the fields that have already
			// been autocompleted, then suggest the closing array bracket ]
			if lo.ContainsBy(allFields, func(f *CompletionItem) bool {
				return f.Label == t.Value()
			}) {
				return []*CompletionItem{{Label: "]", Kind: KindPunctuation}}
			}

			if t.Value() == "]" {
				return []*CompletionItem{{Label: ")", Kind: KindPunctuation}}
			}

			return []*CompletionItem{}
		}

		if enclosingBlock == parser.KeywordFields {
			return []*CompletionItem{}
		}
	}

	return []*CompletionItem{}
}

func getSortableArgCompletions(asts []*parser.AST, t *TokensAtPosition, cfg *config.ProjectConfig) []*CompletionItem {
	modelName := getParentModelName(t)
	model := query.Model(asts, modelName)
	fields := query.ModelFields(model)

	completions := []*CompletionItem{}

	for _, field := range fields {
		fieldName := field.Name.Value

		if query.IsHasOneModelField(asts, field) || query.IsHasManyModelField(asts, field) {
			continue
		}

		completions = append(completions, &CompletionItem{
			Label:       fieldName,
			Description: field.Type.Value,
			Kind:        KindField,
		})
	}

	completions = append(completions, builtInFieldCompletions...)

	return completions
}

func getOrderByArgCompletions(asts []*parser.AST, t *TokensAtPosition, cfg *config.ProjectConfig) []*CompletionItem {

	argStart := t.StartOfParen()
	for {
		if argStart == nil {
			// shouldn't happen but worth being defensive to avoid an infinite loop
			return []*CompletionItem{}
		}
		if argStart.ValueAt(-2) == "@" && argStart.ValueAt(-1) == parser.AttributeOrderBy {
			break
		}
		argStart = argStart.Prev().StartOfParen()
	}

	comma := t.FindPrev(",")

	// This is a big "fudgy" but to detect if the current position is a label
	// we see if the start of the attribute args is the current or previous token
	// or if the current or previous token is a comma...
	isLabel := argStart.Is(t, t.Prev()) || comma.Is(t, t.Prev())

	if isLabel {
		modelName := getParentModelName(t)
		model := query.Model(asts, modelName)
		fields := query.ModelFields(model)

		completions := []*CompletionItem{}

		for _, field := range fields {
			fieldName := field.Name.Value

			if query.IsHasOneModelField(asts, field) || query.IsHasManyModelField(asts, field) {
				continue
			}

			completions = append(completions, &CompletionItem{
				Label:       fieldName,
				Description: field.Type.Value,
				Kind:        KindField,
			})
		}

		completions = append(completions, builtInFieldCompletions...)

		return completions
	}

	return []*CompletionItem{
		{
			Label: "asc",
			Kind:  KindKeyword,
		},
		{
			Label: "desc",
			Kind:  KindKeyword,
		},
	}
}

func getScheduleArgCompletions(asts []*parser.AST, t *TokensAtPosition, cfg *config.ProjectConfig) []*CompletionItem {
	jobs := query.Jobs(asts)

	completions := []*CompletionItem{}

	for _, field := range jobs {
		fieldName := field.Name.Value

		completions = append(completions, &CompletionItem{
			Label:       fieldName,
			Description: field.Name.Value,
			Kind:        KindField,
		})
	}

	completions = append(completions, builtInFieldCompletions...)

	return completions
}

func getPermissionArgCompletions(asts []*parser.AST, t *TokensAtPosition, cfg *config.ProjectConfig) []*CompletionItem {
	argStart := t.StartOfParen()

	for {
		if argStart == nil {
			// shouldn't happen but worth being defensive to avoid an infinite loop
			return []*CompletionItem{}
		}
		if argStart.ValueAt(-2) == "@" && argStart.ValueAt(-1) == parser.AttributePermission {
			break
		}
		argStart = argStart.Prev().StartOfParen()
	}

	colon := t.FindPrev(":")
	comma := t.FindPrev(",")

	listStart := t.StartOfGroup("[", "]")
	// This is a big "fudgy" but to detect if the current position is a label
	// we see if the start of the attribute args is the current or previous token
	// or if the current or previous token is a comma...
	isLabel := argStart.Is(t, t.Prev()) || comma.Is(t, t.Prev())

	// ... however if we are within a list (which can contain commas) then we're
	// not in a label
	if listStart != nil {
		isLabel = false
	}

	// completion for labels
	if isLabel {
		labels := []*CompletionItem{
			{
				Label: "expression",
				Kind:  KindLabel,
			},
			{
				Label: "roles",
				Kind:  KindLabel,
			},
		}
		if getTypeOfEnclosingBlock(t) == parser.KeywordModel {
			labels = append(labels, &CompletionItem{
				Label: "actions",
				Kind:  KindLabel,
			})
		}
		return labels
	}

	// completion for values

	label := colon.Prev().Value()
	switch label {
	case "expression":
		return getExpressionCompletions(asts, t, cfg)
	case "actions":
		if listStart != nil {
			return lo.Filter(actionBlockKeywords, func(c *CompletionItem, _ int) bool {
				return c.Label != parser.KeywordWith
			})
		}
		return []*CompletionItem{}
	case "roles":
		if listStart != nil {
			return getUserDefinedTypeCompletions(asts, t, parser.KeywordRole)
		}
		return []*CompletionItem{}
	default:
		return []*CompletionItem{}
	}
}

func getExpressionCompletions(asts []*parser.AST, t *TokensAtPosition, cfg *config.ProjectConfig) []*CompletionItem {
	modelName := getParentModelName(t)
	expressionModelName := casing.ToLowerCamel(modelName)

	previousIdents := getPreviousIdents(t)
	if len(previousIdents) == 0 {
		return append([]*CompletionItem{
			{
				Label:       casing.ToLowerCamel(modelName),
				Description: "Current model",
				Kind:        KindModel,
			},
			{
				Label:       "ctx",
				Description: "Current context",
				Kind:        KindVariable,
			},
		}, getUserDefinedTypeCompletions(asts, t, "enum")...)
	}

	switch previousIdents[0] {
	case "ctx":
		var completions []*CompletionItem
		completions = []*CompletionItem{
			{
				Label:       "identity",
				Description: "Identity",
				Kind:        KindField,
			},
			{
				Label:       "now",
				Description: "Timestamp",
				Kind:        KindField,
			},
			{
				Label:       "env",
				Description: "Environment Variables",
				Kind:        KindField,
			},
			{
				Label:       "secrets",
				Description: "Secrets",
				Kind:        KindField,
			},
			{
				Label:       "isAuthenticated",
				Description: "Authentication Indicator",
				Kind:        KindField,
			},
			{
				Label:       "headers",
				Description: "Request Headers",
				Kind:        KindField,
			},
		}

		if len(previousIdents) == 2 {
			switch previousIdents[1] {
			case "env":
				completions = getEnvironmentVariableCompletions(cfg)
			case "secrets":
				completions = getSecretsCompletions(cfg)
			}
		}

		return completions

	case expressionModelName:
		model := query.Model(asts, modelName)
		fieldNames, ok := getFieldNamesAtPath(asts, model, previousIdents[1:])
		if !ok {
			// if we were unable to resolve the relevant model
			// return no completions as returning the default
			// fields in this case could be unhelpful
			return []*CompletionItem{}
		}

		fieldNames = append(fieldNames, builtInFieldCompletions...)
		return fieldNames

	default:
		// Enum value completions
		e := query.Enum(asts, previousIdents[0])
		if e != nil {
			return lo.Map(e.Values, func(v *parser.EnumValueNode, _ int) *CompletionItem {
				return &CompletionItem{
					Label: v.Name.Value,
					Kind:  KindField,
				}
			})
		}

		return []*CompletionItem{}
	}
}

var namedBlocks = []string{
	parser.KeywordModel,
	parser.KeywordEnum,
	parser.KeywordApi,
	parser.KeywordRole,
	parser.KeywordMessage,
	parser.KeywordJob,
}

var unNamedBlocks = []string{
	parser.KeywordFields,
	parser.KeywordActions,
	parser.KeywordEmails,
	parser.KeywordDomains,
	parser.KeywordModels,
	parser.KeywordInput,
}

// getTypeOfEnclosingBlock returns the keyword used to define
// the closest parent block or an empty string if not inside
// a block e.g. at the top-level of the schema
func getTypeOfEnclosingBlock(t *TokensAtPosition) string {
	for {
		t = t.StartOfBlock()
		if t == nil {
			break
		}

		for _, k := range unNamedBlocks {
			if k == t.ValueAt(-1) {
				return k
			}
		}

		for _, k := range namedBlocks {
			if k == t.ValueAt(-2) {
				return k
			}
		}

		t = t.Prev()
	}

	return ""
}

// getParentModel name returns the user-defined name of the
// model that `t` is within or an empty string if `t` is not
// contained inside a model
func getParentModelName(t *TokensAtPosition) string {
	for {
		t = t.StartOfBlock()
		if t == nil {
			break
		}

		if t.ValueAt(-2) == parser.KeywordModel {
			return t.ValueAt(-1)
		}

		t = t.Prev()
	}

	return ""
}

var identRegex = regexp.MustCompile("^[a-zA-Z]+$")

// getPreviousIdents returns the idents that are part of the same
// operand e.g. for "foo.bar.baz" if `tokens` was pointed at "baz"
// then this function would return ["foo", "bar"]
func getPreviousIdents(tokens *TokensAtPosition) []string {
	idents := []string{}

	if tokens.tokens[tokens.tokenIndex].Value == "(" {
		return []string{}
	}

	// walk backwards from current token collecting all the idents
	t := tokens
	for {
		t = t.Prev()
		isIdent := identRegex.MatchString(t.Value())
		if !isIdent && t.Value() != "." {
			break
		}
		if isIdent {
			idents = append(idents, t.Value())
		}
	}

	// reverse the idents so they are in the right order
	return lo.Reverse(idents)
}

// getParentAttribute returns the name of the parent attribute
// and true if `t` is within an attributes arguments
// e.g. @set(person.<Cursor) <-- this would return "set", true
// It returns an empty string and false in all other cases including
// that `t` is pointing to the attribute name or the "@" token
func getParentAttribute(t *TokensAtPosition) (string, bool) {
	for {
		if t.Value() == "(" && t.ValueAt(-2) == "@" {
			return t.ValueAt(-1), true
		}
		t = t.Prev().StartOfParen()
		if t == nil {
			return "", false
		}
	}
}

// getUserDefinedTypeCompletions returns the names of user-defined
// types defined with `keyword`
func getUserDefinedTypeCompletions(asts []*parser.AST, t *TokensAtPosition, keyword string) []*CompletionItem {
	completions := []*CompletionItem{}

	// First we take values from the AST's
	switch keyword {
	case "model":
		for _, m := range query.Models(asts) {
			completions = append(completions, &CompletionItem{
				Label: m.Name.Value,
				Kind:  keyword,
			})
		}
	case "enum":
		for _, m := range query.Enums(asts) {
			completions = append(completions, &CompletionItem{
				Label: m.Name.Value,
				Kind:  keyword,
			})
		}
	case "message":
		for _, name := range query.MessageNames(asts) {
			completions = append(completions, &CompletionItem{
				Label: name,
				Kind:  keyword,
			})
		}
	case "job":
		for _, job := range query.Jobs(asts) {
			completions = append(completions, &CompletionItem{
				Label: job.Name.Value,
				Kind:  keyword,
			})
		}
	}

	// Then we inspect the tokens from the active file to find definitions
	// This approach means we can extract definitions that occur after a syntax error
	t = t.Start()
	for {
		if t.Value() == keyword {
			completions = append(completions, &CompletionItem{
				Label: t.ValueAt(1),
				Kind:  keyword,
			})
		}

		t = t.
			Next().       // to the name of the block
			Next().       // to the opening {
			EndOfBlock(). // to the end of the }
			Next()        // to the next token after the }

		if t == nil {
			break
		}
	}

	// Finally make sure there are no duplicates
	return lo.UniqBy(completions, func(r *CompletionItem) string {
		return r.Label
	})
}

func getFieldNamesAtPath(asts []*parser.AST, model *parser.ModelNode, idents []string) ([]*CompletionItem, bool) {
	if model == nil {
		return nil, false
	}

	for i := 0; i < len(idents); i++ {
		ident := idents[i]
		field := query.ModelField(model, ident)
		if field == nil {
			return nil, false
		}
		model = query.Model(asts, field.Type.Value)
		if model == nil {
			return nil, false
		}
	}

	return getModelFieldCompletions(model), true
}

func getModelFieldCompletions(model *parser.ModelNode) []*CompletionItem {
	completions := []*CompletionItem{}
	for _, field := range query.ModelFields(model) {
		completions = append(completions, &CompletionItem{
			Label:       field.Name.Value,
			Description: field.Type.Value,
			Kind:        KindField,
		})
	}

	return completions
}

// getAttributeCompletions returns CompletionItems for each attribute name
// provided in `names`
func getAttributeCompletions(token *TokensAtPosition, names []string) []*CompletionItem {
	completions := []*CompletionItem{}
	for _, v := range names {
		// By default we only insert the name of the attribute. This is
		// becaue the "@" is actually a different token
		insertText := v

		if token.Value() == "" || token.Value() == "@" || strings.HasPrefix(insertText, token.Value()) {

			// The exception is if the current token is whitespace, then we
			// can insert both the "@" and the attribute name
			if token.Value() == "" {
				insertText = "@" + v
			}

			completions = append(completions, &CompletionItem{
				Label:      "@" + v,
				InsertText: insertText,
				Kind:       KindAttribute,
			})
		}
	}
	return completions
}

func getEnvironmentVariableCompletions(cfg *config.ProjectConfig) []*CompletionItem {
	var builtInFieldCompletions []*CompletionItem
	for _, key := range cfg.AllEnvironmentVariables() {
		builtInFieldCompletions = append(builtInFieldCompletions, &CompletionItem{
			Label:       key,
			Description: "Environment Variables",
			Kind:        KindField,
		})

	}
	return builtInFieldCompletions
}

func getSecretsCompletions(cfg *config.ProjectConfig) []*CompletionItem {
	var builtInFieldCompletions []*CompletionItem
	for _, key := range cfg.AllSecrets() {
		builtInFieldCompletions = append(builtInFieldCompletions, &CompletionItem{
			Label:       key,
			Description: "Secret",
			Kind:        KindField,
		})

	}
	return builtInFieldCompletions
}
