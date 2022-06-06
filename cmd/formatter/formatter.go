package formatter

<<<<<<<< HEAD:output/formatter.go
import "bytes"

type FormatterType string
========
type FormatType string
>>>>>>>> @{-1}:cmd/formatter/formatter.go

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
