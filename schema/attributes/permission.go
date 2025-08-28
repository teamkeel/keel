package attributes

import (
	"encoding/hex"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/expressions/typing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var permissions = make(map[string]*expressions.Parser)
var mutex sync.Mutex

// defaultPermission will cache the base CEL environment for a schema.
func defaultPermission(schema []*parser.AST) (*expressions.Parser, error) {
	mutex.Lock()
	defer mutex.Unlock()

	var contents string
	for _, s := range schema {
		contents += s.Raw + "\n"
	}
	key := hex.EncodeToString([]byte(contents))

	if parser, exists := permissions[key]; exists {
		return parser, nil
	}

	opts := []expressions.Option{
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithComparisonOperators(),
		options.WithLogicalOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	}

	parser, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	permissions[key] = parser

	return parser, nil
}

func ValidatePermissionExpression(schema []*parser.AST, entity parser.Entity, action *parser.ActionNode, job *parser.JobNode, expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	parser, err := defaultPermission(schema)
	if err != nil {
		return nil, err
	}

	opts := []expressions.Option{}

	operands, _ := resolve.IdentOperands(expression)

	if action != nil && operands != nil {
	out:
		for _, operand := range operands {
			for _, input := range action.Inputs {
				if operand.Fragments[0] == input.Name() {
					opts = append(opts, options.WithActionInputs(schema, action))
					break out
				}
			}
		}
	}

	if entity != nil {
		for _, operand := range operands {
			if operand.Fragments[0] == strcase.ToLowerCamel(entity.GetName()) {
				opts = append(opts, options.WithVariable(strcase.ToLowerCamel(entity.GetName()), entity.GetName(), false))
				break
			}
		}
	}

	// If there are no options to add, we can just validate without extending the parser
	// This is a performance optimization
	if len(opts) == 0 {
		return parser.Validate(expression)
	}

	p, err := parser.Extend(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}

func ValidatePermissionRoles(schema []*parser.AST, expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	opts := []expressions.Option{
		options.WithSchemaTypes(schema),
		options.WithReturnTypeAssertion(typing.Role.String(), true),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}

func ValidatePermissionActions(expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	opts := []expressions.Option{
		options.WithConstant(parser.ActionTypeGet, "_ActionType"),
		options.WithConstant(parser.ActionTypeCreate, "_ActionType"),
		options.WithConstant(parser.ActionTypeUpdate, "_ActionType"),
		options.WithConstant(parser.ActionTypeList, "_ActionType"),
		options.WithConstant(parser.ActionTypeDelete, "_ActionType"),
		options.WithReturnTypeAssertion("_ActionType", true),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}
