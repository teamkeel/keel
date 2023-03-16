package node

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// For every custom function, the method will translate the contents of any relevant permission expressions into JavaScript code that will infer the necessary joins and constraints to add in order to query the database via Kysely to retrieve the values of left and right hand side operands in each condition of an expression.
// This works for infinitely nested conditions with ands/ors
// For example, given the expression "post.author.title == "123"
// The following Kysely code will be generated:
//
//	  const operand_1 = db.selectFrom("post")
//			.innerJoin("author", "author.id", "post.author_id")
//			.where("post.id", "in", records.map((r) => r.id))
//			.select("author.title as v")
//	   .execute();
//
// And the return statement will look like:
//
// return operand_1.every(x => operand_2.every(y => y === x));
func GeneratePermissionFunctions(w *Writer, schema *proto.Schema) {
	customFns := proto.FilterOperations(schema, func(op *proto.Operation) bool {
		return op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM
	})

	permissionMap := map[string]*proto.PermissionRule{}

	permObjectWriter := Writer{}

	functionNameCounter := 1

	permObjectWriter.Writeln("const permissionFns = {")
	permObjectWriter.Indent()

	for _, action := range customFns {
		permissions := proto.PermissionsForAction(schema, action)
		permissionFunctions := []string{}

		for _, permission := range permissions {
			identifier := fmt.Sprintf("permissionFn_%d", functionNameCounter)
			permissionMap[identifier] = permission
			permissionFunctions = append(permissionFunctions, identifier)
			functionNameCounter++
		}

		permObjectWriter.Writef("%s: [%s],\n", action.Name, strings.Join(permissionFunctions, ", "))
	}

	permObjectWriter.Dedent()
	permObjectWriter.Writeln("}")

	for identifier, permissionRule := range permissionMap {
		expressionString := permissionRule.Expression.Source
		w.Writef("// @permission(expression: %s)\n", expressionString)
		w.Writef("const %s = async (records, ctx, db) => {\n", identifier)
		w.Indent()

		expression, err := parser.ParseExpression(expressionString)

		if err != nil {
			return
		}
		var operandCounter int = 1
		w.Writef("return %s;", generateExpression(w, &operandCounter, schema, expression))

		w.Dedent()
		w.Writeln("")
		w.Writeln("};")
	}

	w.Write(permObjectWriter.String())

	w.Writeln("module.exports.permissionFns = permissionFns;")
}

// generateExpression is responsible for two things:
// 1. Constructing database queries using Kysely that retrieve the value of each operand from the database. If the operand is a literal then its value is just assigned to a variable.
// 2. Building up the return statement that returns true/false depending on the outcome of the *whole* expression
func generateExpression(w *Writer, counter *int, schema *proto.Schema, expression *parser.Expression) string {
	expWriter := Writer{}

	for i, or := range expression.Or {
		if len(or.And) > 1 {
			expWriter.Write("(")
		}
		for j, and := range or.And {
			if and.Condition != nil {
				condition := and.Condition

				lhsIdentifier := writeOperandSql(w, counter, schema, condition.LHS)
				operator := condition.Operator

				// rhs could be nil if its a value expression
				if condition.RHS != nil {
					rhsIdentifier := writeOperandSql(w, counter, schema, condition.RHS)
					// a.every(x => b.every(y => y {operator} x))
					expWriter.Writef(
						`%s.every(x => %s.every(y => y %s x))`,
						lhsIdentifier,
						rhsIdentifier,
						KeelToJSOperatorMap[operator.Symbol],
					)
				} else {
					expWriter.Writef("%s.every(r => r)", lhsIdentifier)
				}
			}

			if j+1 < len(or.And) {
				expWriter.Write(" && ")
			}

			// for nested conditions, we just call this very function recursively to build up the groupings.
			if and.Expression != nil {
				expWriter.Write(generateExpression(w, counter, schema, and.Expression))
			}
		}

		if len(or.And) > 1 {
			expWriter.Write(")")
		}

		if i+1 < len(expression.Or) {
			expWriter.Write(" || ")
		}
	}

	return expWriter.String()
}

type Join struct {
	FromTable  string
	ToTable    string
	ToColumn   string
	FromColumn string
}

// writeOperandSql is responsible for outputting variables to the main writer that hold the values of database queries, literal values and ctx values
func writeOperandSql(w *Writer, counter *int, schema *proto.Schema, operand *parser.Operand) string {
	w.Writef("// operand: %s\n", operand.ToString())

	identifier := fmt.Sprintf("operand_%d", *counter)

	switch {
	case operand.Ident != nil && operand.Ident.IsContext():
		symbol := operand.Ident.ToString()

		if operand.Ident.ToString() == "ctx.identity" {
			symbol = "ctx.identity.id"
		}
		w.Writef("const %s = [%s];\n", identifier, symbol)
	case operand.Ident != nil:
		// if the operand is an ident, then we know we need to construct a Kysely query to retrieve the value from the database.

		rootTable := strcase.ToSnake(operand.Ident.Fragments[0].Fragment)
		field := operand.Ident.LastFragment()

		w.Writef("let %s = await db", identifier)

		w.Indent()
		w.Writef(".selectFrom(\"%s\")\n", rootTable)

		joins := buildJoins(schema, operand.Ident.Fragments)

		for _, join := range joins {
			w.Writef(".innerJoin('%s', '%s', '%s')\n", join.ToTable, fmt.Sprintf("%s.%s", join.ToTable, join.ToColumn), fmt.Sprintf("%s.%s", join.FromTable, join.FromColumn))
		}

		w.Writef(".where('%s', 'in', records.map((r) => r.id))\n", fmt.Sprintf("%s.id", rootTable))

		// If the fragments end with a model e.g ctx.identity, then we append "id" on the end when selecting from the db
		if IsFragmentTerminatingWithModel(schema, operand.Ident.Fragments) {
			w.Writef(".select('%s as v')\n", fmt.Sprintf("%s.id", operand.Ident.Fragments[len(operand.Ident.Fragments)-1].Fragment))
		} else {
			w.Writef(".select('%s as v')\n", fmt.Sprintf("%s.%s", operand.Ident.Fragments[len(operand.Ident.Fragments)-2].Fragment, strcase.ToSnake(field)))
		}
		w.Writeln(".execute();\n")

		w.Dedent()

		// we're only interested in the value of v for each row.
		w.Writef("%s = %s.map(x => x.v);\n", identifier, identifier)
	default:
		// literal
		w.Writef("const %s = [%s];\n", identifier, operand.ToString())
	}

	*counter++

	return identifier
}

// buildJoins takes a fragment list such as post.author.publisher.name and figures out the list of joins required to fetch the terminating value
// So in the example above, we know we need to join from post to publisher via author, so an array of joins will be returned allowing us to do that.
func buildJoins(schema *proto.Schema, identFragments []*parser.IdentFragment) (joins []*Join) {
	fragments := lo.Map(identFragments, func(i *parser.IdentFragment, _ int) string {
		return i.Fragment
	})
	model := strcase.ToCamel(fragments[0])
	fragmentCount := len(fragments)

	for i := 1; i < fragmentCount; i++ {
		currentFragment := fragments[i]

		// if the fragments end with a model rather than a field, then handle this special case
		if IsFragmentTerminatingWithModel(schema, identFragments) && i == fragmentCount-1 {
			relatedModelField := proto.FindField(schema.Models, model, currentFragment)

			joins = append(joins, &Join{
				FromTable:  strcase.ToSnake(model),
				FromColumn: strcase.ToSnake(relatedModelField.ForeignKeyFieldName.Value),
				ToTable:    strcase.ToSnake(relatedModelField.Type.ModelName.Value),
				ToColumn:   "id",
			})

			return joins
		}

		if i < fragmentCount-1 {
			relatedModelField := proto.FindField(schema.Models, model, currentFragment)

			if proto.IsBelongsTo(relatedModelField) {
				// foreign key is on this model
				joins = append(joins, &Join{
					FromTable:  strcase.ToSnake(model),
					FromColumn: strcase.ToSnake(relatedModelField.ForeignKeyFieldName.Value),
					ToTable:    strcase.ToSnake(relatedModelField.Type.ModelName.Value),
					ToColumn:   "id",
				})
			} else {
				// foreign key is on other side
				joins = append(joins, &Join{
					FromTable:  strcase.ToSnake(model),
					FromColumn: "id",
					ToTable:    strcase.ToSnake(relatedModelField.Type.ModelName.Value),
					ToColumn:   fmt.Sprintf("%s_id", strcase.ToSnake(model)),
				})
			}

			model = relatedModelField.Type.ModelName.Value
		}
	}

	return joins
}

func IsFragmentTerminatingWithModel(schema *proto.Schema, identFragments []*parser.IdentFragment) bool {
	fragments := lo.Map(identFragments, func(i *parser.IdentFragment, _ int) string {
		return i.Fragment
	})

	fragmentCount := len(fragments)

	model := strcase.ToCamel(fragments[0])

	lastFragmentIsModel := false

	// start at index 1 as we dont care about the root model
	for i := 1; i < fragmentCount; i++ {
		currentFragment := fragments[i]

		field := proto.FindField(schema.Models, model, currentFragment)

		if field.Type.ModelName != nil {
			model = field.Type.ModelName.Value
		}

		if field.Type.ModelName != nil && i == fragmentCount-1 {
			lastFragmentIsModel = true
		}
	}

	return lastFragmentIsModel
}

var (
	KeelToJSOperatorMap = map[string]string{
		"==": "===", // triple equals
		">":  ">",
		">=": ">=",
		"<":  "<",
		"<=": "<=",
		"!=": "!==",
		// todo: support 'in' and 'not in'
	}
)
