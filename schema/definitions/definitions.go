package definitions

import (
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/reader"
)

type FunctionDefinition struct {
	Name string `json:"name"`
}

type Definition struct {
	Schema   *Position           `json:"schema"`
	Function *FunctionDefinition `json:"function"`
}

type Position struct {
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

func GetDefinition(schemaFiles []*reader.SchemaFile, pos Position) *Definition {
	asts := []*parser.AST{}
	for _, f := range schemaFiles {
		ast, _ := parser.Parse(f)
		if ast != nil {
			asts = append(asts, ast)
		}
	}

	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {
			def := definitionFromField(asts, field, pos)
			if def != nil {
				return def
			}
		}

		for _, fn := range query.ModelActions(model, func(a *parser.ActionNode) bool { return a.IsFunction() }) {
			if !tokenContainsPosition(fn.Name.Tokens[0], pos) {
				continue
			}

			return &Definition{
				Function: &FunctionDefinition{
					Name: fn.Name.Value,
				},
			}
		}

		for _, action := range query.ModelActions(model) {
			for _, input := range action.Inputs {
				def := definitionFromIdent(asts, model, &input.Type, pos)
				if def != nil {
					return def
				}
			}

			for _, input := range action.With {
				def := definitionFromIdent(asts, model, &input.Type, pos)
				if def != nil {
					return def
				}
			}
		}
	}

	return nil
}

func definitionFromField(asts []*parser.AST, field *parser.FieldNode, pos Position) *Definition {
	// The second token is the field type
	tok := field.Tokens[1]

	if !tokenContainsPosition(tok, pos) {
		return nil
	}

	model := query.Model(asts, tok.Value)
	if model != nil {
		return definitionFromPosition(model.Name.Pos)
	}

	enum := query.Enum(asts, tok.Value)
	if enum != nil {
		return definitionFromPosition(enum.Name.Pos)
	}

	return nil
}

func definitionFromIdent(asts []*parser.AST, model *parser.ModelNode, ident *parser.Ident, pos Position) *Definition {
	var field *parser.FieldNode
	for _, i := range ident.Fragments {
		if model == nil {
			break
		}

		field = model.Field(i.Fragment)
		if field == nil {
			break
		}

		model = query.Model(asts, field.Type.Value)
		if !tokenContainsPosition(i.Tokens[0], pos) {
			continue
		}

		return definitionFromPosition(field.Name.Pos)
	}

	return nil
}

func tokenContainsPosition(tok lexer.Token, pos Position) bool {
	line := tok.Pos.Line
	start := tok.Pos.Column
	end := start + len(tok.Value)

	if tok.Pos.Filename != pos.Filename {
		return false
	}

	return line == pos.Line && start <= pos.Column && end >= pos.Column
}

func definitionFromPosition(p lexer.Position) *Definition {
	return &Definition{
		Schema: &Position{
			Filename: p.Filename,
			Line:     p.Line,
			Column:   p.Column,
		},
	}
}
