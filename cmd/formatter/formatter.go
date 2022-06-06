package formatter

type FormatType string

const (
	FormatJSON FormatType = "json"
	FormatText FormatType = "text"
)

// Formatter is a simple interface which takes bytes and produces formatted output:
type Formatter interface {
	Output(output interface{}) error
}
