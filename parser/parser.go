package parser

import (
	"github.com/alecthomas/participle"
	"github.com/teamkeel/keel/proto"
)

type Schema struct {
	Declarations []*Declaration `@@+`
}

type Declaration struct {
	Model *Model `"model" @@`
}

type Model struct {
	Name     string          `@Ident`
	Sections []*ModelSection `"{" @@* "}"`
}

type ModelSection struct {
	Fields    []*ModelField    `( "fields" "{" @@+ "}"`
	Functions []*ModelFunction `| "functions" "{" @@+ "}"`
	Attribute *Attribute       `| @@)`
}

type ModelField struct {
	Name       string       `@Ident`
	Type       string       `@Ident`
	Repeated   bool         `@( "[" "]" )?`
	Attributes []*Attribute `( "{" @@* "}" )?`
}

type Attribute struct {
	Name      string               `"@" @Ident`
	Arguments []*AttributeArgument `( "(" @@ ( "," @@ )* ")" )?`
}

type AttributeArgument struct {
	Expression *Expression `@@`
	Value      *Value      `| @@`
}

type Value struct {
	True   bool     `  @"true"`
	False  bool     `| @"false"`
	String string   `| @String`
	Ident  []string `| ( @Ident ( "." @Ident )* )`
	Array  []*Value `| "[" @@ ("," @@)* "]"`
}

type Expression struct {
	LHS *Value `@@`
	Op  string `@( "in" | "=" )`
	RHS *Value `@@`
}

type ModelFunction struct {
	Create    bool           `( @"create"`
	Get       bool           `| @"get" )`
	Name      string         `@Ident`
	Arguments []*FunctionArg `"(" @@ ( "," @@ )* ")"`
}

type FunctionArg struct {
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
