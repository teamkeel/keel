package parser

import (
	"text/scanner"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Schema struct {
	Pos lexer.Position

	Declarations []*Declaration `@@+`
}

type Declaration struct {
	Pos lexer.Position

	Model *Model `("model" @@`
	API   *API   `| "api" @@)`
}

type Model struct {
	Pos lexer.Position

	Name     string          `@Ident`
	Sections []*ModelSection `"{" @@* "}"`
}

type ModelSection struct {
	Pos lexer.Position

	Fields     []*ModelField  `( "fields" "{" @@+ "}"`
	Functions  []*ModelAction `| "functions" "{" @@+ "}"`
	Operations []*ModelAction `| "operations" "{" @@+ "}"`
	Attribute  *Attribute     `| @@)`
}

type ModelField struct {
	Pos lexer.Position

	BuiltIn    bool
	Name       string       `@Ident`
	Type       string       `@Ident`
	Repeated   bool         `@( "[" "]" )?`
	Attributes []*Attribute `( "{" @@+ "}" )?`
}

type API struct {
	Pos lexer.Position

	Name     string        `@Ident`
	Sections []*APISection `"{" @@* "}"`
}

type APISection struct {
	Pos lexer.Position

	Models    []*APIModels `("models" "{" @@* "}"`
	Attribute *Attribute   `| @@)`
}

type APIModels struct {
	Pos lexer.Position

	ModelName string `@Ident`
}

type Attribute struct {
	Pos lexer.Position

	Name      string               `"@" @Ident`
	Arguments []*AttributeArgument `( "(" @@ ( "," @@ )* ")" )?`
}

type AttributeArgument struct {
	Pos lexer.Position

	Name       string      `(@Ident ":")?`
	Expression *Expression `( @@`
	Value      *Value      `| @@`
	Array      []*Value    `| "[" @@ ("," @@)* "]" )`
}

type Value struct {
	Pos lexer.Position

	True   bool     `  @"true"`
	False  bool     `| @"false"`
	String string   `| @String`
	Ident  []string `| ( @Ident ( "." @Ident )* )`
}

type Expression struct {
	Pos lexer.Position

	LHS *Value `@@`
	Op  string `@( "=" "=" | "!" "=" | "=" )`
	RHS *Value `@@`
}

type ModelAction struct {
	Pos lexer.Position

	Type       string       `@Ident`
	Name       string       `@Ident`
	Arguments  []*ActionArg `"(" ( @@ ( "," @@ )* )? ")"`
	Attributes []*Attribute `( "{" @@+ "}" )?`
}

type ActionArg struct {
	Pos lexer.Position

	Name string `@Ident`
}

func Parse(s string) (*Schema, error) {

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
	err = parser.ParseString("", s, schema)
	if err != nil {
		return nil, err
	}

	return schema, nil
}
