package output

import "bytes"

type FormatterType string

const (
	FormatterJSON    FormatterType = "json"
	FormatterConsole FormatterType = "console"
)

type Outputter interface {
	Result() bytes.Buffer
}

// Formatter is a simple interface which takes bytes and produces formatted output:
type Formatter interface {
	Output(output []byte) error
}
