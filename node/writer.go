package node

import (
	"fmt"
	"strings"
)

type Writer struct {
	b      strings.Builder
	indent int
}

func (w *Writer) Indent() {
	w.indent += 4
}

func (w *Writer) Dedent() {
	w.indent -= 4
	if w.indent < 0 {
		w.indent = 0
	}
}

func (w *Writer) isStartOfLine() bool {
	s := w.b.String()
	return len(s) > 0 && s[len(s)-1] == '\n'
}

func (w *Writer) Write(v string) {
	if w.isStartOfLine() && v != "" {
		w.b.WriteString(strings.Repeat(" ", w.indent))
	}
	w.b.WriteString(v)
}

func (w *Writer) Writeln(v string) {
	if w.isStartOfLine() && v != "" {
		w.b.WriteString(strings.Repeat(" ", w.indent))
	}
	w.b.WriteString(v + "\n")
}

func (w *Writer) Writef(v string, a ...any) {
	if w.isStartOfLine() {
		w.b.WriteString(strings.Repeat(" ", w.indent))
	}
	w.b.WriteString(fmt.Sprintf(v, a...))
}

func (w *Writer) String() string {
	return w.b.String()
}
