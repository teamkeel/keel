package completions

import (
	"regexp"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
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
	KindModel     = "model"
	KindField     = "field"
	KindVariable  = "variable"
	KindType      = "type"
	KindKeyword   = "keyword"
	KindLabel     = "label"
	KindAttribute = "attribute"
)

func Completions(schema string, pos *node.Position) []*CompletionItem {

	// parse the schema ignoring any errors, it's very likely the
	// schema is not in a valid state
	ast, _ := parser.Parse(&reader.SchemaFile{
		Contents: schema,
	})

	tokenAtPos := NewTokensAtPosition(schema, pos)

	// First check if we're within an attribute's argumens list
	// Attributes can appear in a number of places so easier to check
	// for this up-front
	_, isAttr := getParentAttribute(tokenAtPos)
	if isAttr {
		return getAttributeArgCompletions(ast, tokenAtPos)
	}

	// switch on nearest (previous) keyword
	switch getTypeOfEnclosingBlock(tokenAtPos) {
	case parser.KeywordModel:
		attributes := getAttributeCompletions(tokenAtPos, []string{parser.AttributePermission})
		return append(attributes, modelBlockKeywords...)
	case parser.KeywordRole:
		return roleBlockKeywords
	case parser.KeywordApi:
		attributes := getAttributeCompletions(tokenAtPos, []string{parser.AttributeGraphQL})
		return append(attributes, apiBlockKeywords...)
	case parser.KeywordEnum:
		// no completions for enum block
		return []*CompletionItem{}
	case parser.KeywordFields:
		return getFieldCompletions(ast, tokenAtPos)
	case parser.KeywordOperations, parser.KeywordFunctions:
		return getActionCompletions(ast, tokenAtPos)
	case parser.KeywordModels:
		// models block inside an api block - complete with model names
		return getUserDefinedTypeCompletions(tokenAtPos, parser.KeywordModel)
	default:
		// If no enclosing block then we're at the top-level of the schema

		// if the previous token is one of the top level keywords then the current token
		// is a name and we can't provide any completions for that
		_, ok := lo.Find(topLevelKeywords, func(v *CompletionItem) bool {
			return v.Label == tokenAtPos.ValueAt(-1)
		})
		if ok {
			return []*CompletionItem{}
		}

		return topLevelKeywords
	}
}

func getFieldCompletions(ast *parser.AST, tokenAtPos *TokensAtPosition) []*CompletionItem {
	// First we find the start of the current block
	startOfBlock := tokenAtPos.StartOfBlock()

	// Simple case for field attributes:
	//   1. Current token is "@"
	//   2. Previous token is "@"
	//   3. Parent block is field block which can only contain attributes
	if tokenAtPos.Value() == "@" ||
		tokenAtPos.ValueAt(-1) == "@" ||
		startOfBlock.Prev().Value() != parser.KeywordFields {
		return getAttributeCompletions(tokenAtPos, []string{
			parser.AttributeUnique,
			parser.AttributeDefault,
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
			})
		}

		// We on a new line which means current token is field name for
		// which we can't provide completions
		return []*CompletionItem{}
	}

	// Provide completions for field type which is built-in and user-defined types
	return lo.Flatten(
		[][]*CompletionItem{
			getUserDefinedTypeCompletions(tokenAtPos, parser.KeywordModel),
			getUserDefinedTypeCompletions(tokenAtPos, parser.KeywordEnum),
			getBuiltInTypeCompletions(),
		},
	)
}

func getActionCompletions(ast *parser.AST, tokenAtPos *TokensAtPosition) []*CompletionItem {
	// if we are inside enclosing parenthesis then we are completing for
	// action inputs
	if tokenAtPos.StartOfParen() != nil {
		return getActionInputCompletions(ast, tokenAtPos)
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
		})
	}

	// current token is action name - can't auto-complete
	if lo.Contains(parser.ActionTypes, tokenAtPos.ValueAt(-1)) {
		return []*CompletionItem{}
	}

	// action block keywords
	return actionBlockKeywords
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
		Label: parser.KeywordOperations,
		Kind:  KindKeyword,
	},
	{
		Label: parser.KeywordFunctions,
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

func getActionInputCompletions(ast *parser.AST, tokenAtPos *TokensAtPosition) []*CompletionItem {
	modelName := getParentModelName(tokenAtPos)
	model := query.Model([]*parser.AST{ast}, modelName)

	// inside action input args - auto-complete field names
	completions := append([]*CompletionItem{}, builtInFieldCompletions...)

	// if we have been able to get the model from the AST we can try to
	// find the field names
	if model != nil {

		fieldNames := getModelFieldCompletions(model)

		// if current or previous token is a "." then we need to provide
		// completions for a related model
		if tokenAtPos.Value() == "." || tokenAtPos.ValueAt(-1) == "." {
			idents := getPreviousIdents(tokenAtPos)
			var ok bool
			fieldNames, ok = getFieldNamesAtPath(ast, model, idents)
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

func getAttributeArgCompletions(ast *parser.AST, t *TokensAtPosition) []*CompletionItem {
	attrName, _ := getParentAttribute(t)

	switch attrName {
	case parser.AttributeSet, parser.AttributeWhere, parser.AttributeValidate:
		return getExpressionCompletions(ast, t)
	case parser.AttributePermission:
		return getPermissionArgCompletions(ast, t)
	default:
		// this is likely a user error e.g. providing args to an attribute
		// that doesn't take them like @unique
		return []*CompletionItem{}
	}
}

func getPermissionArgCompletions(ast *parser.AST, t *TokensAtPosition) []*CompletionItem {
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
		return getExpressionCompletions(ast, t)
	case "actions":
		if listStart != nil {
			return lo.Filter(actionBlockKeywords, func(c *CompletionItem, _ int) bool {
				return c.Label != parser.KeywordWith
			})
		}
		return []*CompletionItem{}
	case "roles":
		if listStart != nil {
			return getUserDefinedTypeCompletions(t, parser.KeywordRole)
		}
		return []*CompletionItem{}
	default:
		return []*CompletionItem{}
	}
}

func getExpressionCompletions(ast *parser.AST, t *TokensAtPosition) []*CompletionItem {
	modelName := getParentModelName(t)
	expressionModelName := strcase.ToLowerCamel(modelName)

	previousIdents := getPreviousIdents(t)
	if len(previousIdents) == 0 {
		return []*CompletionItem{
			{
				Label:       strcase.ToLowerCamel(modelName),
				Description: "Current model",
				Kind:        KindModel,
			},
			{
				Label:       "ctx",
				Description: "Current context",
				Kind:        KindVariable,
			},
		}
	}

	switch previousIdents[0] {
	case "ctx":
		return []*CompletionItem{
			{Label: "identity", Description: "Identity", Kind: KindField},
			{Label: "now", Description: "Timestamp", Kind: KindField},
		}

	case expressionModelName:
		model := query.Model([]*parser.AST{ast}, modelName)
		fieldNames, ok := getFieldNamesAtPath(ast, model, previousIdents[1:])
		if !ok {
			// if we were unable to resolve the relevant model
			// return no completions as returning the default
			// fields in this case could be unhelpful
			return []*CompletionItem{}
		}

		fieldNames = append(fieldNames, builtInFieldCompletions...)
		return fieldNames

	default:
		// TODO: enumm and action inputs
		return []*CompletionItem{}
	}
}

var namedBlocks = []string{
	parser.KeywordModel,
	parser.KeywordEnum,
	parser.KeywordApi,
	parser.KeywordRole,
}

var unNamedBlocks = []string{
	parser.KeywordFields,
	parser.KeywordOperations,
	parser.KeywordFunctions,
	parser.KeywordEmails,
	parser.KeywordDomains,
	parser.KeywordModels,
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
func getUserDefinedTypeCompletions(t *TokensAtPosition, keyword string) []*CompletionItem {
	completions := []*CompletionItem{}

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

	return completions
}

func getFieldNamesAtPath(ast *parser.AST, model *parser.ModelNode, idents []string) ([]*CompletionItem, bool) {
	if model == nil {
		return nil, false
	}

	for i := 0; i < len(idents); i++ {
		ident := idents[i]
		field := query.ModelField(model, ident)
		if field == nil {
			return nil, false
		}
		model = query.Model([]*parser.AST{ast}, field.Type)
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
			Description: field.Type,
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
	return completions
}
