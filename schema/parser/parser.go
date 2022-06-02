package parser

import (
	"text/scanner"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/teamkeel/keel/model"
	"github.com/teamkeel/keel/schema/expressions"
)

type Schema struct {
	Node

	Declarations []*Declaration `@@+`
}

type Declaration struct {
	Node

	Model *Model `("model" @@`
	Role  *Role  `| "role" @@`
	API   *API   `| "api" @@)`
}

type Model struct {
	Node

	Name     NameToken       `@@`
	Sections []*ModelSection `"{" @@* "}"`
}

type ModelSection struct {
	Node

	Fields     []*ModelField  `( "fields" "{" @@+ "}"`
	Functions  []*ModelAction `| "functions" "{" @@+ "}"`
	Operations []*ModelAction `| "operations" "{" @@+ "}"`
	Attribute  *Attribute     `| @@)`
}

type NameToken struct {
	Node

	Text string `@Ident`
}

type AttributeNameToken struct {
	Node

	Text string `"@" @Ident`
}

type ModelField struct {
	Node

	BuiltIn    bool
	Name       NameToken    `@@`
	Type       string       `@Ident`
	Repeated   bool         `@( "[" "]" )?`
	Attributes []*Attribute `( "{" @@+ "}" )?`
}

type API struct {
	Node

	Name     NameToken     `@@`
	Sections []*APISection `"{" @@* "}"`
}

type Role struct {
	Node

	Name     NameToken      `@@`
	Sections []*RoleSection `"{" @@* "}"`
}

type RoleSection struct {
	Node

	Domains []*RoleDomain `("domains" "{" @@* "}"`
	Emails  []*RoleEmail  `| "emails" "{" @@* "}")`
}

type RoleDomain struct {
	Node

	Domain string `@String`
}

type RoleEmail struct {
	Node

	Email string `@String`
}

type APISection struct {
	Node

	Models    []*APIModels `("models" "{" @@* "}"`
	Attribute *Attribute   `| @@)`
}

type APIModels struct {
	Node

	Name NameToken `@@`
}

type Attribute struct {
	Node

	Name      AttributeNameToken   `@@`
	Arguments []*AttributeArgument `( "(" @@ ( "," @@ )* ")" )?`
}

type AttributeArgument struct {
	Node

	Name       NameToken               `(@@ ":")?`
	Expression *expressions.Expression `@@`
}

type ModelAction struct {
	Node

	Type       string       `@Ident`
	Name       NameToken    `@@`
	Arguments  []*ActionArg `"(" ( @@ ( "," @@ )* )? ")"`
	Attributes []*Attribute `( "{" @@+ "}" )?`
}

type ActionArg struct {
	Node

	Name NameToken `@@`
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

func Parse(s *model.SchemaFile) (*Schema, error) {
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

	parser, err := participle.Build(&Schema{}, participle.Lexer(lex))
	if err != nil {
		return nil, err
	}

	schema := &Schema{}

	err = parser.ParseString(s.FileName, s.Contents, schema)
	if err != nil {
		return nil, err
	}

	return schema, nil
}
