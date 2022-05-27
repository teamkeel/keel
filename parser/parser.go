package parser

import (
	"text/scanner"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/model"
)

type Schema struct {
	Pos    lexer.Position
	EndPos lexer.Position

	Declarations []*Declaration `@@+`
}

type Declaration struct {
	Pos    lexer.Position
	EndPos lexer.Position

	Model *Model `("model" @@`
	Role  *Role  `| "role" @@`
	API   *API   `| "api" @@)`
}

type Model struct {
	Pos    lexer.Position
	EndPos lexer.Position

	NameToken NameToken       `@@`
	Sections  []*ModelSection `"{" @@* "}"`
}

type ModelSection struct {
	Pos    lexer.Position
	EndPos lexer.Position

	Fields     []*ModelField  `( "fields" "{" @@+ "}"`
	Functions  []*ModelAction `| "functions" "{" @@+ "}"`
	Operations []*ModelAction `| "operations" "{" @@+ "}"`
	Attribute  *Attribute     `| @@)`
}

type NameToken struct {
	Pos    lexer.Position
	EndPos lexer.Position

	Name string `@Ident`
}

type AttributeNameToken struct {
	Pos    lexer.Position
	EndPos lexer.Position

	Name string `"@" @Ident`
}

type ModelField struct {
	Pos    lexer.Position
	EndPos lexer.Position

	BuiltIn    bool
	NameToken  NameToken    `@@`
	Type       string       `@Ident`
	Repeated   bool         `@( "[" "]" )?`
	Attributes []*Attribute `( "{" @@+ "}" )?`
}

type API struct {
	Pos    lexer.Position
	EndPos lexer.Position

	NameToken NameToken     `@@`
	Sections  []*APISection `"{" @@* "}"`
}

type Role struct {
	Pos    lexer.Position
	EndPos lexer.Position

	NameToken NameToken      `@@`
	Sections  []*RoleSection `"{" @@* "}"`
}

type RoleSection struct {
	Pos    lexer.Position
	EndPos lexer.Position

	Domains []*RoleDomain `("domains" "{" @@* "}"`
	Emails  []*RoleEmail  `| "emails" "{" @@* "}")`
}

type RoleDomain struct {
	Pos    lexer.Position
	EndPos lexer.Position

	Domain string `@String`
}

type RoleEmail struct {
	Pos    lexer.Position
	EndPos lexer.Position

	Email string `@String`
}

type APISection struct {
	Pos    lexer.Position
	EndPos lexer.Position

	Models    []*APIModels `("models" "{" @@* "}"`
	Attribute *Attribute   `| @@)`
}

type APIModels struct {
	Pos    lexer.Position
	EndPos lexer.Position

	ModelNameToken NameToken `@@`
}

type Attribute struct {
	Pos    lexer.Position
	EndPos lexer.Position

	NameToken AttributeNameToken   `@@`
	Arguments []*AttributeArgument `( "(" @@ ( "," @@ )* ")" )?`
}

type AttributeArgument struct {
	Pos    lexer.Position
	EndPos lexer.Position

	NameToken  NameToken               `(@@ ":")?`
	Expression *expressions.Expression `@@`
}

type ModelAction struct {
	Pos    lexer.Position
	EndPos lexer.Position

	Type       string       `@Ident`
	NameToken  NameToken    `@@`
	Arguments  []*ActionArg `"(" ( @@ ( "," @@ )* )? ")"`
	Attributes []*Attribute `( "{" @@+ "}" )?`
}

type ActionArg struct {
	Pos    lexer.Position
	EndPos lexer.Position

	NameToken NameToken `@@`
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
	// TODO: pass filename as first argument
	err = parser.ParseString(s.FileName, s.Contents, schema)
	if err != nil {
		return nil, err
	}

	return schema, nil
}
