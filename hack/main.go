package main

import (
	"fmt"
	"strings"
	"text/scanner"
)

const (
	KindIdent  = "Ident"
	KindBang   = "!"
	KindParenL = "("
	KindParenR = ")"
	KindBraceL = "{"
	KindBraceR = "}"
	KindColon  = ":"
	KindEquals = "="

	// BANG
	// DOLLAR
	// PAREN_L
	// PAREN_R
	// SPREAD
	// COLON
	// EQUALS
	// AT
	// BRACKET_L
	// BRACKET_R
	// BRACE_L
	// PIPE
	// BRACE_R
	// NAME
	// INT
	// FLOAT
	// STRING
	// BLOCK_STRING
	// AMP
)

type Position struct {
	Line     int
	Column   int
	Offset   string
	Filename string
}

type Token struct {
	Kind  string
	Value string
	Pos   Position
}

func main() {
	const src = `model Person
    fields {
        name Text
    }
}`

	var s scanner.Scanner
	s.Init(strings.NewReader(src))
	s.Mode = scanner.ScanIdents |
		scanner.ScanFloats |
		scanner.ScanChars |
		scanner.ScanStrings |
		scanner.ScanRawStrings |
		scanner.ScanComments
	s.Filename = "example"

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		fmt.Printf("%s: %s\n", s.Position, s.TokenText())
	}

}
