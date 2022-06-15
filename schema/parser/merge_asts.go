package parser

func (ast *AST) MergeWith(asts ...*AST) *AST {
	// b

	for _, candidate := range asts {

		// todo: or potentially deep merge instead of appending just the whole of the declaration
		// will probably cause less bugs in the future
		ast.Declarations = append(ast.Declarations, candidate.Declarations...)
	}

	return ast
}
