package output

import (
	"io"

	"github.com/teamkeel/keel/output/console"
	"github.com/teamkeel/keel/output/json"
)

var defaultFormatterFunc = json.New

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
func (o *Output) SetOutput(formatterType FormatterType, writer io.Writer) {
	switch formatterType {
	case FormatterJSON:
		o.formatter = json.New(writer)
	default:
		o.formatter = console.New(writer)
	}
}

// Write uses the current output formatter to write out the interface provided:
func (o *Output) Write(output []byte) error {
	return o.formatter.Output(output)
}
