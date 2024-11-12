package parser

import (
	"fmt"
	"strings"
	"text/scanner"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/reader"
)

type AST struct {
	node.Node

	Declarations         []*DeclarationNode `@@*`
	EnvironmentVariables []string
	Secrets              []string
}

type DeclarationNode struct {
	node.Node

	Model   *ModelNode   `("model" @@`
	Role    *RoleNode    `| "role" @@`
	API     *APINode     `| "api" @@`
	Enum    *EnumNode    `| "enum" @@`
	Message *MessageNode `| "message" @@`
	Job     *JobNode     `| "job" @@)`
}

type ModelNode struct {
	node.Node

	Name     NameNode            `@@`
	Sections []*ModelSectionNode `"{" @@* "}"`
	BuiltIn  bool
}

type ModelSectionNode struct {
	node.Node

	Fields    []*FieldNode   `( ("fields" "{" @@* "}")`
	Actions   []*ActionNode  `| ("actions" "{" @@* "}")`
	Attribute *AttributeNode `| @@ )`
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
	Type       NameNode         `@@`
	Repeated   bool             `@( "[" "]" )?`
	Optional   bool             `@( "?" )?`
	Attributes []*AttributeNode `( "{" @@+ "}" | @@+ )?`

	// Some fields are added implicitly after parsing the schema
	// For these fields this value is set to true so we can distinguish
	// them from fields defined by the user in the schema
	BuiltIn bool
}

func (f *FieldNode) IsScalar() bool {
	switch f.Type.Value {
	case
		FieldTypeBoolean,
		FieldTypeNumber,
		FieldTypeDecimal,
		FieldTypeText,
		FieldTypeDatetime,
		FieldTypeDate,
		FieldTypeSecret,
		FieldTypeID,
		FieldTypePassword,
		FieldTypeMarkdown,
		FieldTypeVector:
		return true
	default:
		return false
	}
}

type APINode struct {
	node.Node

	Name     NameNode          `@@`
	Sections []*APISectionNode `"{" @@* "}"`
}

type APISectionNode struct {
	node.Node

	Models    []*APIModelNode `( "models" "{" @@* "}"`
	Attribute *AttributeNode  `| @@ )`
}

type APIModelNode struct {
	node.Node

	Name     NameNode               `@@`
	Sections []*APIModelSectionNode `("{" @@* "}")*`
}

type APIModelSectionNode struct {
	node.Node

	Actions []*APIModelActionNode `"actions" "{" @@* "}"`
}

type APIModelActionNode struct {
	node.Node

	Name NameNode `@@`
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

type JobNode struct {
	node.Node

	Name     NameNode          `@@`
	Sections []*JobSectionNode `"{" @@* "}"`
}

type JobSectionNode struct {
	node.Node

	Inputs    []*JobInputNode `( "inputs" "{" @@* "}"`
	Attribute *AttributeNode  `| @@ )`
}

type JobInputNode struct {
	node.Node

	Name     NameNode `@@`
	Type     NameNode `@@`
	Repeated bool     `( @( "[" "]" )`
	Optional bool     `| @( "?" ))?`
}

// type AttributeNode interface {
// 	value()
// 	Type() string
// }

// Attributes:
// - @permission
// - @set
// - @validate
// - @where
// - @unique
// - @default
// - @orderBy
// - @sortable
// // - @on
type AttributeNode struct {
	node.Node

	Name AttributeNameToken `@@`

	// This supports:
	// - no parenthesis at all
	// - empty parenthesis
	// - parenthesis with args
	Arguments []*AttributeArgumentNode `(( "(" @@ ( "," @@ )* ")" ) | ( "(" ")" ) )?`
}

// type PermissionAttributeNode struct {
// 	node.Node

// 	Arguments []*AttributeArgumentNode `"@" "permission" "(" @@ ( "," @@ )* ")"`
// }

// func (PermissionAttributeNode) value()       {}
// func (PermissionAttributeNode) Type() string { return AttributePermission }

// type WhereAttributeNode struct {
// 	node.Node
// 	//Arguments string `"@" "where" "(" @Ident ")"`
// 	Arguments *AttributeArgumentNode `"@" "where" "(" @@ ")"`
// }

// func (WhereAttributeNode) value()       {}
// func (WhereAttributeNode) Type() string { return AttributeWhere }

// type SetAttributeNode struct {
// 	node.Node
// 	//Arguments string `"@" "where" "(" @Ident ")"`
// 	Arguments *AttributeArgumentNode `"@" "set" "(" @@ ")"`
// }

// func (SetAttributeNode) value()       {}
// func (SetAttributeNode) Type() string { return AttributeSet }

// type OrderByAttributeNode struct {
// 	node.Node
// 	//Arguments string `"@" "where" "(" @Ident ")"`
// 	Arguments []*AttributeArgumentNode `"@" "orderBy" "(" @@ ( "," @@ )* ")"`
// }

// func (OrderByAttributeNode) value()       {}
// func (OrderByAttributeNode) Type() string { return AttributeSet }

type AttributeArgumentNode struct {
	node.Node

	Label      *NameNode   `(@@ ":")?`
	Expression *Expression `@@`
}

type ExpressionNode struct {
	node.Node

	Value *string
}

func (e *ExpressionNode) Operands() {

}

type Grouped struct {
	node.Node

	Exp *Expression `"(" @@ ")"`
}

type Expression struct {
	node.Node

	LHS      *Term       `@@`
	Operator *Operator   `( @@`
	RHS      *Expression `@@ )?`
}

type Term struct {
	node.Node

	LHS      *Factor   `@@`
	Operator *Operator `( @@`
	RHS      *Term     `  @@ )?`
}

type Factor struct {
	node.Node

	Grouped  *Grouped  `  @@`
	Function *Function `| @@`
	Operand  *Operand  `| @@`
}

type Function struct {
	node.Node

	Name      *NameNode     `@@`
	Arguments []*Expression `"(" (@@ ( "," @@ )* )? ")"`
}

type ExpressionPart interface {
	String() string
	Operands() []*Operand
}

func (e *Expression) String() string {
	if e.Operator == nil || e.RHS == nil {
		return e.LHS.String()
	}
	return fmt.Sprintf("%s %s %s", e.LHS.String(), e.Operator.ToString(), e.RHS.String())
}

func (t *Term) String() string {
	if t.Operator == nil || t.RHS == nil {
		return t.LHS.String()
	}
	return fmt.Sprintf("%s %s %s", t.LHS.String(), t.Operator.ToString(), t.RHS.String())
}

func (f *Factor) String() string {
	switch {
	case f.Grouped != nil:
		return f.Grouped.String()
	case f.Function != nil:
		return f.Function.String()
	case f.Operand != nil:
		return f.Operand.ToString()
	default:
		return ""
	}
}

func (g *Grouped) String() string {
	if g.Exp == nil {
		return "()"
	}
	return fmt.Sprintf("(%s)", g.Exp.String())
}

func (f *Function) String() string {
	args := make([]string, len(f.Arguments))
	for i, arg := range f.Arguments {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", f.Name.Value, strings.Join(args, ", "))
}

func (e *Expression) Operands() []*Operand {
	if e.Operator == nil || e.RHS == nil {
		return e.LHS.Operands()
	}
	return append(e.LHS.Operands(), e.RHS.Operands()...)
}

func (t *Term) Operands() []*Operand {
	if t.Operator == nil || t.RHS == nil {
		return t.LHS.Operands()
	}
	return append(t.LHS.Operands(), t.RHS.Operands()...)
}

func (f *Factor) Operands() []*Operand {
	switch {
	case f.Grouped != nil:
		return f.Grouped.Operands()
	case f.Function != nil:
		return f.Function.Operands()
	case f.Operand != nil:
		return []*Operand{f.Operand}
	default:
		return []*Operand{}
	}
}

func (g *Grouped) Operands() []*Operand {
	if g.Exp == nil {
		return []*Operand{}
	}
	return g.Exp.Operands()
}

func (f *Function) Operands() []*Operand {
	args := make([]*Operand, len(f.Arguments))
	for _, arg := range f.Arguments {
		args = append(args, arg.Operands()...)
	}
	return args
}

// type ExpressionNode struct {
// 	node.Node

// 	Value  *string
// 	Tokens []*lexer.Token
// }

// func (b *ExpressionNode) Parse(lex *lexer.PeekingLexer) error {
// 	expression := ""
// 	tokens := []*lexer.Token{}

// 	for {
// 		token := lex.Peek()

// 		if token.Value == ")" || token.Value == "," { // TODO: terribly crude
// 			break
// 		}

// 		expression += lex.Next().Value
// 		tokens = append(tokens, token)
// 	}

// 	*b = ExpressionNode{
// 		Value: &expression,
// 	}

// 	return nil
// }

type ActionNode struct {
	node.Node

	Type       NameNode           `@@`
	Name       NameNode           `@@`
	Inputs     []*ActionInputNode `"(" ( @@ ( "," @@ )* ","? )? ")"`
	With       []*ActionInputNode `( ( "with" "(" ( @@ ( "," @@ )* ","? )? ")" )`
	Returns    []*ActionInputNode `| ( "returns" "(" ( @@ ( "," @@ )* ) ")" ) )?`
	Attributes []*AttributeNode   `( "{" @@+ "}" | @@+ )?`

	BuiltIn bool
}

func (a *ActionNode) IsArbitraryFunction() bool {
	return a.IsFunction() && (a.Type.Value == ActionTypeRead || a.Type.Value == ActionTypeWrite)
}

func (a *ActionNode) IsFunction() bool {
	if a.Type.Value == ActionTypeRead || a.Type.Value == ActionTypeWrite {
		return true
	}
	return lo.ContainsBy(a.Attributes, func(a *AttributeNode) bool {
		return a.Name.Value == AttributeFunction
	})
}

type ActionInputNode struct {
	node.Node

	Label    *NameNode `(@@ ":")?`
	Type     Ident     `@@`
	Optional bool      `@( "?" )?`
}

func (a *ActionInputNode) Name() string {
	if a.Label != nil {
		return a.Label.Value
	}

	// if label is not provided then it's computed from the type
	// e.g. if type is `post.author.name` then the input is called `postAuthorName`
	builder := strings.Builder{}
	for _, frag := range a.Type.Fragments {
		builder.WriteString(casing.ToCamel(frag.Fragment))
	}

	return casing.ToLowerCamel(builder.String())
}

type EnumNode struct {
	node.Node

	Name   NameNode         `@@`
	Values []*EnumValueNode `"{" @@* "}"`
}

type EnumValueNode struct {
	node.Node

	Name NameNode `@@`
}

type MessageNode struct {
	node.Node

	Name    NameNode     `@@`
	Fields  []*FieldNode `"{" @@* "}"`
	BuiltIn bool
}

func (e *EnumNode) NameNode() NameNode {
	return e.Name
}

func (e *MessageNode) NameNode() NameNode {
	return e.Name
}

func (e *ModelNode) NameNode() NameNode {
	return e.Name
}

type Error struct {
	err participle.Error
}

// compile-time check that Error inplements node.ParserNode
var _ node.ParserNode = Error{}

func (e Error) Error() string {
	msg := e.err.Error()
	pos := e.err.Position()

	// error messages start with "{filename}:{line}:{column}:" and we don't
	// really need that bit so we can remove it
	return strings.TrimPrefix(msg, fmt.Sprintf("%s:%d:%d:", pos.Filename, pos.Line, pos.Column))
}

func (e Error) GetPositionRange() (start lexer.Position, end lexer.Position) {
	pos := e.err.Position()
	return pos, pos
}

func (e Error) InRange(position node.Position) bool {
	// Just use Node's implementation of InRange
	return node.Node{Pos: e.err.Position()}.InRange(position)
}

func (e Error) HasEndPosition() bool {
	// Just use Node's implementation of HasEndPosition
	return node.Node{Pos: e.err.Position()}.HasEndPosition()
}

func (e Error) GetTokens() []lexer.Token {
	return []lexer.Token{}
}

func Parse(s *reader.SchemaFile) (*AST, error) {
	// Customise the lexer to not ignore comments
	lex := lexer.NewTextScannerLexer(func(s *scanner.Scanner) {
		s.Mode =
			scanner.ScanIdents |
				scanner.ScanFloats |
				scanner.ScanChars |
				scanner.ScanStrings |
				scanner.ScanComments
	})

	parser, err := participle.Build[AST](
		participle.Lexer(lex),
		//participle.Union[AttributeNode](WhereAttributeNode{}, PermissionAttributeNode{}),
		//	participle.Union[AbsExpression](BinaryExpression{}, Operand{}, FunctionCall{}, Grouped{}),
		participle.Elide("Comment"))
	if err != nil {
		return nil, err
	}

	schema, err := parser.ParseString(s.FileName, s.Contents)
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
