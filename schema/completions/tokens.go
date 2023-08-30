package completions

import (
	"strings"
	"text/scanner"

	"github.com/teamkeel/keel/schema/node"
)

// Token represents a single token e.g. "@" or "model"
type Token struct {
	Value string
	Pos   *node.Position
}

// TokensAtPosition represents a list of tokens with a current position
// The position should be considered immutable for each TokensAtPosition
// instance.
type TokensAtPosition struct {
	tokens     []*Token
	tokenIndex int
}

func NewTokensAtPosition(schema string, pos *node.Position) *TokensAtPosition {
	var s scanner.Scanner
	s.Init(strings.NewReader(schema))
	s.Filename = ""

	tokens := &TokensAtPosition{
		tokenIndex: -1,
	}

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		t := &Token{
			Value: s.TokenText(),
			Pos: &node.Position{
				Line:   s.Pos().Line,
				Column: s.Pos().Column,
			},
		}

		tokenStart := t.Pos.Column - len(t.Value)

		// If token index is not set but this token comes after the given
		// position then insert a whitespace token
		if tokens.tokenIndex == -1 && (t.Pos.Line > pos.Line || t.Pos.Line == pos.Line && tokenStart > pos.Column) {
			tokens.tokens = append(tokens.tokens, &Token{
				Value: "",
				Pos: &node.Position{
					Column: pos.Column,
					Line:   pos.Line,
				},
			})
			tokens.tokenIndex = len(tokens.tokens) - 1
		}

		tokens.tokens = append(tokens.tokens, t)

		// Set current token index if this token matches position
		if t.Pos.Line == pos.Line && tokenStart < pos.Column && t.Pos.Column >= pos.Column {
			tokens.tokenIndex = len(tokens.tokens) - 1
		}
	}

	return tokens
}

func (t *TokensAtPosition) Value() string {
	if t == nil || len(t.tokens) == 0 {
		return ""
	}
	return t.tokens[t.tokenIndex].Value
}

func (t *TokensAtPosition) Line() int {
	if t == nil || len(t.tokens) == 0 {
		return -1
	}
	return t.tokens[t.tokenIndex].Pos.Line
}

func (t *TokensAtPosition) LineAt(offset int) int {
	if t == nil {
		return -1
	}
	i := t.tokenIndex + offset
	if i < 0 || i > len(t.tokens)-1 {
		return -1
	}

	return t.tokens[i].Pos.Line
}

func (t *TokensAtPosition) ValueAt(offset int) string {
	if t == nil {
		return ""
	}
	i := t.tokenIndex + offset
	if i < 0 || i > len(t.tokens)-1 {
		return ""
	}

	return t.tokens[i].Value
}

func (t *TokensAtPosition) Next() *TokensAtPosition {
	if t == nil {
		return nil
	}
	if t.tokenIndex >= len(t.tokens)-1 {
		return nil
	}

	return &TokensAtPosition{
		tokens:     t.tokens,
		tokenIndex: t.tokenIndex + 1,
	}
}

func (t *TokensAtPosition) Start() *TokensAtPosition {
	return &TokensAtPosition{
		tokens:     t.tokens,
		tokenIndex: 0,
	}
}

func (t *TokensAtPosition) Prev() *TokensAtPosition {
	if t == nil {
		return nil
	}
	if t.tokenIndex == 0 {
		return nil
	}

	return &TokensAtPosition{
		tokens:     t.tokens,
		tokenIndex: t.tokenIndex - 1,
	}
}

func (t *TokensAtPosition) FindPrev(value string) *TokensAtPosition {
	for {
		if t.Value() == value {
			return t
		}
		t = t.Prev()
		if t == nil {
			return nil
		}
	}
}

func (t *TokensAtPosition) FindPrevMultiple(values ...string) (string, *TokensAtPosition) {
	for {
		for _, value := range values {
			if t.Value() == value {
				return value, t
			}
		}

		t = t.Prev()
		if t == nil {
			return "", nil
		}
	}
}

func (t *TokensAtPosition) FindPrevMultipleOnLine(values ...string) (string, *TokensAtPosition) {
	currentToken := t
	for {
		for _, value := range values {
			if t.Value() == value {
				return value, t
			}
		}

		t = t.Prev()
		if t == nil || currentToken.Line() > t.Line() {
			return "", nil
		}
	}
}

/**
* Find a previous token with a given value on the same line as the current token
**/
func (t *TokensAtPosition) FindPrevOnLine(value string) *TokensAtPosition {
	prev := t.FindPrev(value)
	if prev.Line() == t.Line() {
		return prev
	}
	return nil
}

/**
* Find a previous token with a given value on the same line the given offset
**/
func (t *TokensAtPosition) FindPrevOnLineAt(value string, offset int) *TokensAtPosition {
	prev := t.FindPrev(value)
	if prev.Line() == offset {
		return prev
	}
	return nil
}

func (t *TokensAtPosition) Is(others ...*TokensAtPosition) bool {
	if t == nil {
		return false
	}
	for _, other := range others {
		if other == nil {
			continue
		}
		posA := t.tokens[t.tokenIndex].Pos
		posB := other.tokens[other.tokenIndex].Pos
		if posA.Line == posB.Line && posA.Column == posB.Column {
			return true
		}
	}

	return false
}

/**
* Returns true of the current token is at the beginning of a new line
**/
func (t *TokensAtPosition) IsNewLine() bool {
	return t.Line() > t.Prev().Line()
}

func (t *TokensAtPosition) StartOfBlock() *TokensAtPosition {
	return t.StartOfGroup("{", "}")
}

func (t *TokensAtPosition) StartOfParen() *TokensAtPosition {
	return t.StartOfGroup("(", ")")
}

func (t *TokensAtPosition) EndOfBlock() *TokensAtPosition {
	return t.EndOfGroup("{", "}")
}

func (t *TokensAtPosition) EndOfParen() *TokensAtPosition {
	return t.EndOfGroup("(", ")")
}

func (t *TokensAtPosition) StartOfGroup(start string, end string) *TokensAtPosition {
	if t == nil {
		return nil
	}

	count := 0
	for i := t.tokenIndex; i >= 0; i = i - 1 {
		token := t.tokens[i]

		if token.Value == start {
			count--

			// if we're still waiting to close a group then continue
			if count >= 0 {
				continue
			}

			return &TokensAtPosition{
				tokens:     t.tokens,
				tokenIndex: i,
			}
		}

		// If we find an end token that isn't the token we started on increment a counter
		if token.Value == end && i != t.tokenIndex {
			count++
			continue
		}
	}

	return nil
}

func (t *TokensAtPosition) EndOfGroup(start string, end string) *TokensAtPosition {
	count := 0
	for i := t.tokenIndex; i < len(t.tokens); i = i + 1 {
		token := t.tokens[i]

		if token.Value == end {
			count--

			// if we're still waiting to close a group then continue
			if count >= 0 {
				continue
			}

			return &TokensAtPosition{
				tokens:     t.tokens,
				tokenIndex: i,
			}
		}

		// If we find a start token that isn't the token we started on increment a counter
		if token.Value == start && i != t.tokenIndex {
			count++
			continue
		}
	}

	return nil
}
