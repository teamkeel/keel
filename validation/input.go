package validation

import "github.com/teamkeel/keel/parser"

type Input struct{
	FileName string
	ParsedSchema *parser.Schema
}