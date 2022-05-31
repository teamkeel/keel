package formatter

import (
	"io"
)

var defaultFormatterFunc = NewTextFormatter

// Output is a configurable output mechanism:
type Output struct {
	formatter Formatter
}

// New returns a default output:
func New(writer io.Writer) *Output {
	return &Output{
		formatter: defaultFormatterFunc(writer),
	}
}

// SetOutput allows us to change the output at runtime:
func (o *Output) SetOutput(formatterType FormatType, writer io.Writer) {
	switch formatterType {
	case FormatJSON:
		o.formatter = NewJSONFormatter(writer)
	default:
		o.formatter = NewTextFormatter(writer)
	}
}

// Write uses the current output formatter to write out the interface provided:
func (o *Output) Write(output interface{}) error {
	return o.formatter.Output(output)
}
