package parser

import (
	"text/scanner"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/proto"
)

type Schema struct {
	Pos lexer.Position

	Declarations []*Declaration `@@+`
}

type Declaration struct {
	Pos lexer.Position

	Model *Model `("model" @@`
	Role  *Role  `| "role" @@`
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

type Role struct {
	Pos lexer.Position

	Name     string         `@Ident`
	Sections []*RoleSection `"{" @@* "}"`
}

type RoleSection struct {
	Pos lexer.Position

	Domains []*RoleDomain `("domains" "{" @@* "}"`
	Emails  []*RoleEmail  `| "emails" "{" @@* "}")`
}

type RoleDomain struct {
	Pos lexer.Position

	Domain string `@String`
}

type RoleEmail struct {
	Pos lexer.Position

	Email string `@String`
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

	Name       string                  `(@Ident ":")?`
	Expression *expressions.Expression `@@`
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

func ToProto(s *Schema) (*proto.Schema, error) {
	ps := &proto.Schema{}

	for _, dec := range s.Declarations {
		if dec.Model == nil {
			continue
		}

		m := &proto.Model{
			Name: dec.Model.Name,
		}

		for _, sec := range dec.Model.Sections {
			if sec.Fields == nil {
				continue
			}
			for _, field := range sec.Fields {
				f := &proto.Field{
					Name: field.Name,
				}

				m.Fields = append(m.Fields, f)
			}
		}
	}

	return ps, nil
}
