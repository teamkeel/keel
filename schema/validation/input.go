package validation

import "github.com/teamkeel/keel/schema/parser"

type Input struct {
	FileName     string
	ParsedSchema *parser.AST
}
