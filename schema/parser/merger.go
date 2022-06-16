package parser

func (ast *AST) MergeWith(asts ...*AST) (res *AST) {
	// b

	res = ast

	// for _, candidate := range asts {

	// 	// todo: or potentially deep merge instead of appending just the whole of the declaration
	// 	// will probably cause less bugs in the future
	// 	// ast.Declarations = append(ast.Declarations, candidate.Declarations...)

	// 	for _, decl := range candidate.Declarations {
	// 		for _, modelSection := range decl.Model.Sections {
	// 			target := query.Model(ast, decl.Model.Name.Value)
	// 		}
	// 	}

	// 	// deep merge:
	// 	// models
	// 	//   actions (funcs/ops) - stay dupe
	// 	//   fields - just append
	// 	//   append attrs
	// 	// api
	// 	//   models - append
	// 	//   append attrs
	// 	// role (same name)
	// 	//   append emails
	// 	//   append domains

	// }

	return ast
}
