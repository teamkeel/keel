package parser

import (
	"fmt"
	"strings"
	"text/scanner"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/reader"
)

type AST struct {
	node.Node

	Declarations []*DeclarationNode `@@+`
}

type DeclarationNode struct {
	node.Node

	Model *ModelNode `("model" @@`
	Role  *RoleNode  `| "role" @@`
	API   *APINode   `| "api" @@)`
	Enum  *EnumNode  `| @@`
}

type ModelNode struct {
	node.Node

	Name     NameNode            `@@`
	Sections []*ModelSectionNode `"{" @@* "}"`
	BuiltIn  bool
}

func (model *ModelNode) ToString() string {
	return model.Name.Value
}

type ModelSectionNode struct {
	node.Node

	Fields     []*FieldNode   `( "fields" "{" @@+ "}"`
	Functions  []*ActionNode  `| "functions" "{" @@+ "}"`
	Operations []*ActionNode  `| "operations" "{" @@+ "}"`
	Attribute  *AttributeNode `| @@)`
}

type NameNode struct {
	node.Node

	Value string `@Ident`
}

type AttributeNameToken struct {
	node.Node

	Value string `"@" @Ident`
}

type FieldNode struct {
	node.Node

	Name       NameNode         `@@`
	Type       string           `@Ident`
	Repeated   bool             `( @( "[" "]" )`
	Optional   bool             `| @( "?" ))?`
	Attributes []*AttributeNode `( "{" @@+ "}" | @@+ )?`

	// Some fields are added implicitly after parsing the schema
	// For these fields this value is set to true so we can distinguish
	// them from fields defined by the user in the schema
	BuiltIn bool
}

func (field *FieldNode) ToString() string {
	return field.Name.Value
}

type APINode struct {
	node.Node

	Name     NameNode          `@@`
	Sections []*APISectionNode `"{" @@* "}"`
}

type APISectionNode struct {
	node.Node

	Models    []*ModelsNode  `("models" "{" @@* "}"`
	Attribute *AttributeNode `| @@)`
}

type RoleNode struct {
	node.Node

	Name     NameNode           `@@`
	Sections []*RoleSectionNode `"{" @@* "}"`
}

type RoleSectionNode struct {
	node.Node

	Domains []*DomainNode `("domains" "{" @@* "}"`
	Emails  []*EmailsNode `| "emails" "{" @@* "}")`
}

type DomainNode struct {
	node.Node

	Domain string `@String`
}

type EmailsNode struct {
	node.Node

	Email string `@String`
}

type ModelsNode struct {
	node.Node

	Name NameNode `@@`
}

type AttributeNode struct {
	node.Node

	Name      AttributeNameToken       `@@`
	Arguments []*AttributeArgumentNode `( "(" @@ ( "," @@ )* ")" )?`
}

type AttributeArgumentNode struct {
	node.Node

	Label      *NameNode               `(@@ ":")?`
	Expression *expressions.Expression `@@`
}

type ActionNode struct {
	node.Node

	Type       string             `@Ident`
	Name       NameNode           `@@`
	Inputs     []*ActionInputNode `"(" ( @@ ( "," @@ )* )? ")"`
	With       []*ActionInputNode `( "with" "(" ( @@ ( "," @@ )* ) ")" )?`
	Attributes []*AttributeNode   `( "{" @@+ "}" )?`
}

type ActionInputNode struct {
	node.Node

	Label    *NameNode         `(@@ ":")?`
	Type     expressions.Ident `@@`
	Repeated bool              `( @( "[" "]" )`
	Optional bool              `| @( "?" ))?`
}

func (a *ActionInputNode) Name() string {
	if a.Label != nil {
		return a.Label.Value
	}

	frags := a.Type.Fragments

	// if label is not provided then it's computed from the type
	// e.g. if type is `post.author.name` then the input is called `authorName`
	name := frags[len(frags)-1].Fragment
	if len(frags) > 1 {
		name = frags[len(frags)-2].Fragment + strcase.ToCamel(name)
	}

	return name
}

type EnumNode struct {
	node.Node

	Name   NameNode         `"enum" @@`
	Values []*EnumValueNode `"{" @@+ "}"`
}

type EnumValueNode struct {
	node.Node

	Name NameNode `@@`
}

type Error struct {
	err participle.Error
}

func (e Error) Error() string {
	msg := e.err.Error()
	pos := e.err.Position()

	// error messages start with "{filename}:{line}:{column}:" and we don't
	// really need that bit so we can remove it
	return strings.TrimPrefix(msg, fmt.Sprintf("%s:%d:%d:", pos.Filename, pos.Line, pos.Column))
}

// Implement node.Node interface
func (e Error) GetPositionRange() (start lexer.Position, end lexer.Position) {
	pos := e.err.Position()
	return pos, pos
}

func Parse(s *reader.SchemaFile) (*AST, error) {
	// Customise the lexer to not ignore comments
	lex := lexer.NewTextScannerLexer(func(s *scanner.Scanner) {
		s.Mode =
			scanner.ScanIdents |
				scanner.ScanFloats |
				scanner.ScanChars |
				scanner.ScanStrings |
				scanner.ScanRawStrings |
				scanner.ScanComments
	})

	parser, err := participle.Build(&AST{}, participle.Lexer(lex))
	if err != nil {
		return nil, err
	}

	schema := &AST{}

	err = parser.ParseString(s.FileName, s.Contents, schema)
	if err != nil {

		// If the error is a participle.Error (which it should be)
		// then return an error that also implements the node.Node
		// interface so that we can later on turn it into a validation
		// error
		perr, ok := err.(participle.Error)
		if ok {
			return nil, Error{perr}
		}

		return nil, err
	}

	return schema, nil
}
