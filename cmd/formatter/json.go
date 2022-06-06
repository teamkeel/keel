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
<<<<<<<< HEAD:output/json/formatter.go
func (f *Formatter) Output(output []byte) error {
========
func (f *JSONFormatter) Output(output interface{}) error {
>>>>>>>> @{-1}:cmd/formatter/json.go
	return f.encoder.Encode(output)
}
