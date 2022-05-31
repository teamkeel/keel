package json

import (
	"encoding/json"
	"io"
)

// Formatter will output as formatted JSON:
type Formatter struct {
	encoder *json.Encoder
}

// New returns a JSON formatter:
func New(writer io.Writer) *Formatter {
	newEncoder := json.NewEncoder(writer)
	newEncoder.SetIndent("", "    ")

	return &Formatter{
		encoder: newEncoder,
	}
}

// Output implements the Formatter interface:
func (f *Formatter) Output(output []byte) error {
	return f.encoder.Encode(output)
}
