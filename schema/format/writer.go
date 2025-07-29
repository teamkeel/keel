package format

import (
	"fmt"
	"strings"
	"text/scanner"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/node"
)

type Writer struct {
	b          strings.Builder
	currIndent int

	// We keep a stack of comments as when ending a block
	// we will print any trailing comments inside the closing
	// paren
	commentStack [][]lexer.Token

	// We keep a cache of which comments we've already printed
	// as the same comment tokens can appear on different nodes
	commentCache map[string]bool
}

func (w *Writer) WriteLine(s string) {
	if w.IsStartOfLine() && s != "" {
		w.b.WriteString(strings.Repeat(" ", w.currIndent))
	}
	w.b.WriteString(fmt.Sprintf(s + "\n"))
}

func (w *Writer) Write(s string, args ...any) {
	if w.IsStartOfLine() && s != "" {
		w.b.WriteString(strings.Repeat(" ", w.currIndent))
	}
	w.b.WriteString(fmt.Sprintf(s, args...))
}

func (w *Writer) Indent() {
	w.currIndent += indentSize
}

func (w *Writer) Dedent() {
	w.currIndent -= indentSize
	if w.currIndent < 0 {
		w.currIndent = 0
	}
}

func (w *Writer) Block(fn func()) {
	w.WriteLine(" {")
	w.Indent()
	fn()
	if len(w.commentStack) > 0 {
		tokens := w.commentStack[len(w.commentStack)-1]
		w.TrailingComments(tokens)
	}
	w.Dedent()
	w.WriteLine("}")
}

func (w *Writer) LineLength() int {
	s := w.b.String()
	lines := strings.Split(s, "\n")
	curr := lines[len(lines)-1]
	return len(curr)
}

func (w *Writer) Comments(node node.ParserNode, fn func()) {
	tokens := node.GetTokens()
	w.commentStack = append(w.commentStack, tokens)

	w.LeadingComments(tokens)
	fn()
	w.TrailingComments(tokens)

	w.commentStack = w.commentStack[0 : len(w.commentStack)-1]
}

func (w *Writer) LeadingComments(tokens []lexer.Token) {
	for _, t := range tokens {
		if t.Type != scanner.Comment {
			return
		}
		if !w.SeenToken(t) {
			w.WriteLine(t.Value)
		}
	}
}

func (w *Writer) TrailingComments(tokens []lexer.Token) {
	comments := []lexer.Token{}
	for i := len(tokens) - 1; i >= 0; i-- {
		t := tokens[i]
		if t.Type == '}' {
			continue
		}
		if t.Type != scanner.Comment {
			break
		}
		comments = append(comments, t)
	}
	for _, t := range lo.Reverse(comments) {
		if !w.SeenToken(t) {
			w.WriteLine(t.Value)
		}
	}
}

func (w *Writer) SeenToken(t lexer.Token) bool {
	key := fmt.Sprintf("%d:%d", t.Pos.Line, t.Pos.Column)
	_, seen := w.commentCache[key]
	if !seen {
		w.commentCache[key] = true
	}
	return seen
}

func (w *Writer) String() string {
	return w.b.String()
}

func (w *Writer) IsStartOfLine() bool {
	s := w.b.String()
	return len(s) > 0 && s[len(s)-1] == '\n'
}
