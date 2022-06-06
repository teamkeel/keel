package formatter

import (
	"fmt"
	"io"
	"strings"
)

// TextFormatter will output as formatted JSON:
type TextFormatter struct {
	printer io.Writer
}

// New returns a console formatter:
func NewTextFormatter(writer io.Writer) *TextFormatter {
	return &TextFormatter{
		printer: writer,
	}
}

// Output implements the Formatter interface:
func (f *TextFormatter) Output(output interface{}) error {
	var s string

	switch v := output.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		s = fmt.Sprintf("%v", v)
	}

	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}

	_, err := f.printer.Write([]byte(s))
	return err
}
