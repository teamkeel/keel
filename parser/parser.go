package parser

import (
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/teamkeel/keel/proto"
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

	Name       string       `@Ident`
	Type       string       `@Ident`
	Repeated   bool         `@( "[" "]" )?`
	Attributes []*Attribute `( "{" @@* "}" )?`
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

	Expression *Expression `@@`
	Value      *Value      `| @@`
}

type Value struct {
	Pos lexer.Position

	True   bool     `  @"true"`
	False  bool     `| @"false"`
	String string   `| @String`
	Ident  []string `| ( @Ident ( "." @Ident )* )`
	Array  []*Value `| "[" @@ ("," @@)* "]"`
}

type Expression struct {
	Pos lexer.Position

	LHS *Value `@@`
	Op  string `@( "in" | "=" )`
	RHS *Value `@@`
}

type ModelAction struct {
	Pos lexer.Position

	Type      string       `@Ident`
	Name      string       `@Ident`
	Arguments []*ActionArg `"(" @@ ( "," @@ )* ")"`
}

type ActionArg struct {
	Pos lexer.Position

	Name string `@Ident`
}

func Parse(s string) (*Schema, error) {
	parser, err := participle.Build(&Schema{}, participle.UseLookahead(3))
	if err != nil {
		return nil, err
	}

	schema := &Schema{}
	err = parser.ParseString(s, schema)
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
