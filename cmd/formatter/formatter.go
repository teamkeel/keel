package formatter

import "bytes"

type FormatterType string

type FormatType string

const (
	FormatJSON FormatType = "json"
	FormatText FormatType = "text"
)

type Outputter interface {
	Result() bytes.Buffer
}

// Formatter is a simple interface which takes bytes and produces formatted output:
type Formatter interface {
	Output(output []byte) error
}
