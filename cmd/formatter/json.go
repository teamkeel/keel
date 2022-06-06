package formatter

import (
	"encoding/json"
	"io"
)

// JSONFormatter will output as formatted JSON:
type JSONFormatter struct {
	encoder *json.Encoder
}

// New returns a JSON formatter:
func NewJSONFormatter(writer io.Writer) *JSONFormatter {
	newEncoder := json.NewEncoder(writer)
	newEncoder.SetIndent("", "    ")

	return &JSONFormatter{
		encoder: newEncoder,
	}
}

// Output implements the Formatter interface:
func (f *JSONFormatter) Output(output interface{}) error {
	return f.encoder.Encode(output)
}
