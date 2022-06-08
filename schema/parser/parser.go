package parser

import (
	"text/scanner"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/reader"
)

type AST struct {
	Node

	Declarations []*DeclarationNode `@@+`
}

type DeclarationNode struct {
	Node

	Model *ModelNode `("model" @@`
	Role  *RoleNode  `| "role" @@`
	API   *APINode   `| "api" @@)`
	Enum  *EnumNode  `| @@`
}

type ModelNode struct {
	Node

	Name     NameNode            `@@`
	Sections []*ModelSectionNode `"{" @@* "}"`
}

type ModelSectionNode struct {
	Node

	Fields     []*FieldNode   `( "fields" "{" @@+ "}"`
	Functions  []*ActionNode  `| "functions" "{" @@+ "}"`
	Operations []*ActionNode  `| "operations" "{" @@+ "}"`
	Attribute  *AttributeNode `| @@)`
}

type NameNode struct {
	Node

	Value string `@Ident`
}

type AttributeNameToken struct {
	Node

	Value string `"@" @Ident`
}

type FieldNode struct {
	Node

	Name       NameNode         `@@`
	Type       string           `@Ident`
	Repeated   bool             `@( "[" "]" )?`
	Attributes []*AttributeNode `( "{" @@+ "}" )?`

	// Some fields are added implicitly after parsing the schema
	// For these fields this value is set to true so we can distinguish
	// them from fields defined by the user in the schema
	BuiltIn bool
}

type APINode struct {
	Node

	Name     NameNode          `@@`
	Sections []*APISectionNode `"{" @@* "}"`
}

type APISectionNode struct {
	Node

	Models    []*ModelsNode  `("models" "{" @@* "}"`
	Attribute *AttributeNode `| @@)`
}

type RoleNode struct {
	Node

	Name     NameNode           `@@`
	Sections []*RoleSectionNode `"{" @@* "}"`
}

type RoleSectionNode struct {
	Node

	Domains []*DomainNode `("domains" "{" @@* "}"`
	Emails  []*EmailsNode `| "emails" "{" @@* "}")`
}

type DomainNode struct {
	Node

	Domain string `@String`
}

type EmailsNode struct {
	Node

	Email string `@String`
}

type ModelsNode struct {
	Node

	Name NameNode `@@`
}

type AttributeNode struct {
	Node

	Name      AttributeNameToken       `@@`
	Arguments []*AttributeArgumentNode `( "(" @@ ( "," @@ )* ")" )?`
}

type AttributeArgumentNode struct {
	Node

	Name       NameNode                `(@@ ":")?`
	Expression *expressions.Expression `@@`
}

type ActionNode struct {
	Node

	Type       string                `@Ident`
	Name       NameNode              `@@`
	Arguments  []*ActionArgumentNode `"(" ( @@ ( "," @@ )* )? ")"`
	Attributes []*AttributeNode      `( "{" @@+ "}" )?`
}

type ActionArgumentNode struct {
	Node

	Name NameNode `@@`
}

type EnumNode struct {
	Node

	Name   NameNode         `"enum" @@`
	Values []*EnumValueNode `"{" @@+ "}"`
}

type EnumValueNode struct {
	Node

	Name NameNode `@@`
}

type Node struct {
	Pos    lexer.Position
	EndPos lexer.Position
	Tokens []lexer.Token
}

// GetPositionRange returns a start and end position that correspond to Node
// The behaviour of start position is exactly the same as the Pos field that
// participle provides but the end position is calculated from the position of
// the last token in this node, which is more useful if you want to know where
// _this_ node starts and ends.
func (n Node) GetPositionRange() (start lexer.Position, end lexer.Position) {
	start.Column = n.Pos.Column
	start.Filename = n.Pos.Filename
	start.Line = n.Pos.Line
	start.Offset = n.Pos.Offset

	// This shouldn't really happen but just to be safe
	if len(n.Tokens) == 0 {
		return start, n.EndPos
	}

	lastToken := n.Tokens[len(n.Tokens)-1]
	endPos := lastToken.Pos

	tokenLength := utf8.RuneCountInString(lastToken.Value)

	end.Filename = endPos.Filename

	// assumption here is that a token can't span multiple lines, which
	// I'm pretty sure is true
	end.Line = endPos.Line

	// Update offset and column to reflect the end of last token
	// in this node
	end.Offset = endPos.Offset + tokenLength
	end.Column = endPos.Column + tokenLength

	return start, end
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
		return nil, err
	}

	return schema, nil
}
