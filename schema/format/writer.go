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

func (w *Writer) writeLine(s string) {
	if w.isStartOfLine() && s != "" {
		w.b.WriteString(strings.Repeat(" ", w.currIndent))
	}
	w.b.WriteString(fmt.Sprintf(s + "\n"))
}

func (w *Writer) write(s string, args ...any) {
	if w.isStartOfLine() && s != "" {
		w.b.WriteString(strings.Repeat(" ", w.currIndent))
	}
	w.b.WriteString(fmt.Sprintf(s, args...))
}

func (w *Writer) indent() {
	w.currIndent += indentSize
}

func (w *Writer) dedent() {
	w.currIndent -= indentSize
	if w.currIndent < 0 {
		w.currIndent = 0
	}
}

func (w *Writer) block(fn func()) {
	w.writeLine(" {")
	w.indent()
	fn()
	if len(w.commentStack) > 0 {
		tokens := w.commentStack[len(w.commentStack)-1]
		w.trailingComments(tokens)
	}
	w.dedent()
	w.writeLine("}")
}

func (w *Writer) lineLength() int {
	s := w.b.String()
	lines := strings.Split(s, "\n")
	curr := lines[len(lines)-1]
	return len(curr)
}

func (w *Writer) comments(node node.ParserNode, fn func()) {
	tokens := node.GetTokens()
	w.commentStack = append(w.commentStack, tokens)

	w.leadingComments(tokens)
	fn()
	w.trailingComments(tokens)

	w.commentStack = w.commentStack[0 : len(w.commentStack)-1]
}

func (w *Writer) leadingComments(tokens []lexer.Token) {
	for _, t := range tokens {
		if t.Type != scanner.Comment {
			return
		}
		if !w.seenToken(t) {
			w.writeLine(t.Value)
		}
	}
}

func (w *Writer) trailingComments(tokens []lexer.Token) {
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
		if !w.seenToken(t) {
			w.writeLine(t.Value)
		}
	}
}

func (w *Writer) seenToken(t lexer.Token) bool {
	key := fmt.Sprintf("%d:%d", t.Pos.Line, t.Pos.Column)
	_, seen := w.commentCache[key]
	if !seen {
		w.commentCache[key] = true
	}
	return seen
}

func (w *Writer) string() string {
	return w.b.String()
}

func (w *Writer) isStartOfLine() bool {
	s := w.b.String()
	return len(s) > 0 && s[len(s)-1] == '\n'
}
