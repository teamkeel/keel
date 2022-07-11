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

type GenericNode interface {
	node.ParserNode
	fmt.Stringer
}

type AST struct {
	node.Node

	Declarations []*DeclarationNode `@@*`
}

func (ast *AST) String() string {
	return "AST"
}

type DeclarationNode struct {
	node.Node

	Model *ModelNode `("model" @@`
	Role  *RoleNode  `| "role" @@`
	API   *APINode   `| "api" @@`
	Enum  *EnumNode  `| "enum" @@)`
}

func (ast *DeclarationNode) String() string {
	return "DeclarationNode"
}

type ModelNode struct {
	node.Node

	Name     NameNode            `@@`
	Sections []*ModelSectionNode `"{" @@* "}"`
	BuiltIn  bool
}

func (ast *ModelNode) String() string {
	return "ModelNode"
}

type ModelSectionNode struct {
	node.Node

	Fields     []*FieldNode   `( "fields" "{" @@? "}"`
	Functions  []*ActionNode  `| "functions" "{" @@* "}"`
	Operations []*ActionNode  `| "operations" "{" @@* "}"`
	Attribute  *AttributeNode `| @@)`
}

func (ast *ModelSectionNode) String() string {
	return "ModelSectionNode"
}

type NameNode struct {
	node.Node

	Value string `@Ident`
}

func (ast *NameNode) String() string {
	return "NameNode"
}

type AttributeNameToken struct {
	node.Node

	Value string `"@" @Ident`
}

func (ast *AttributeNameToken) String() string {
	return "AttributeNameToken"
}

type FieldNode struct {
	node.Node

	Name       NameNode         `@@`
	Type       TypeNode         `@@?`
	Repeated   bool             `( @( "[" "]" )`
	Optional   bool             `| @( "?" ))?`
	Attributes []*AttributeNode `( "{" @@+ "}" | @@+ )?`

	// Some fields are added implicitly after parsing the schema
	// For these fields this value is set to true so we can distinguish
	// them from fields defined by the user in the schema
	BuiltIn bool
}

func (field *FieldNode) String() string {
	return field.Name.Value
}

type TypeNode struct {
	node.Node

	Value string `" " @Ident`
}

func (t *TypeNode) String() string {
	return t.Value
}

type APINode struct {
	node.Node

	Name     NameNode          `@@`
	Sections []*APISectionNode `"{" @@* "}"`
}

func (ast *APINode) String() string {
	return "APINode"
}

type APISectionNode struct {
	node.Node

	Models    []*ModelsNode  `("models" "{" @@* "}"`
	Attribute *AttributeNode `| @@)`
}

func (ast *APISectionNode) String() string {
	return "APISectionNode"
}

type RoleNode struct {
	node.Node

	Name     NameNode           `@@`
	Sections []*RoleSectionNode `"{" @@* "}"`
}

func (ast *RoleNode) String() string {
	return "RoleNode"
}

type RoleSectionNode struct {
	node.Node

	Domains []*DomainNode `("domains" "{" @@* "}"`
	Emails  []*EmailsNode `| "emails" "{" @@* "}")`
}

func (ast *RoleSectionNode) String() string {
	return "RoleSectionNode"
}

type DomainNode struct {
	node.Node

	Domain string `@String`
}

func (ast *DomainNode) String() string {
	return "DomainNode"
}

type EmailsNode struct {
	node.Node

	Email string `@String`
}

func (ast *EmailsNode) String() string {
	return "EmailsNode"
}

type ModelsNode struct {
	node.Node

	Name NameNode `@@`
}

func (ast *ModelsNode) String() string {
	return "ModelsNode"
}

type AttributeNode struct {
	node.Node

	Name      AttributeNameToken       `@@`
	Arguments []*AttributeArgumentNode `( "(" @@ ( "," @@ )* ")" )?`
}

func (ast *AttributeNode) String() string {
	return "AttributeNode"
}

type AttributeArgumentNode struct {
	node.Node

	Label      *NameNode               `(@@ ":")?`
	Expression *expressions.Expression `@@`
}

func (ast *AttributeArgumentNode) String() string {
	return "AttributeArgumentNode"
}

type ActionNode struct {
	node.Node

	Type       string             `@Ident`
	Name       NameNode           `@@`
	Inputs     []*ActionInputNode `"(" ( @@ ( "," @@ )* )? ")"`
	With       []*ActionInputNode `( "with" "(" ( @@ ( "," @@ )* ) ")" )?`
	Attributes []*AttributeNode   `( "{" @@+ "}" )?`
}

func (ast *ActionNode) String() string {
	return "ActionNode"
}

type ActionInputNode struct {
	node.Node

	Label    *NameNode         `(@@ ":")?`
	Type     expressions.Ident `@@`
	Repeated bool              `( @( "[" "]" )`
	Optional bool              `| @( "?" ))?`
}

func (ast *ActionInputNode) String() string {
	return "ActionInputNode"
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

	Name   NameNode         `@@`
	Values []*EnumValueNode `"{" @@* "}"`
}

func (ast *EnumNode) String() string {
	return "EnumNode"
}

type EnumValueNode struct {
	node.Node

	Name NameNode `@@`
}

func (ast *EnumValueNode) String() string {
	return "EnumValueNode"
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

func (e Error) InRange(position node.Position) bool {
	// Just use Node's implementation of InRange
	return node.Node{Pos: e.err.Position()}.InRange(position)
}

func (e Error) GetTokens() []lexer.Token {
	return []lexer.Token{}
}

func (e Error) HasEndPosition() bool {
	// Just use Node's implementation of InRange
	return node.Node{Pos: e.err.Position()}.HasEndPosition()

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
			return schema, Error{perr}
		}

		return schema, err
	}

	return schema, nil
}
