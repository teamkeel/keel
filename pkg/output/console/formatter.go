package console

import (
	"fmt"
	"io"
)

// Formatter will output as formatted JSON:
type Formatter struct {
	printer io.Writer
}

// New returns a console formatter:
func New(writer io.Writer) *Formatter {
	return &Formatter{
		printer: writer,
	}
}

// Output implements the Formatter interface:
func (f *Formatter) Output(output interface{}) error {
	if _, err := f.printer.Write([]byte(fmt.Sprintf("%v", output))); err != nil {
		return err
	}
	return nil
}
