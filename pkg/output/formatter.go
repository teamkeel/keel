package output

type FormatterType string

const (
	FormatterJSON    FormatterType = "json"
	FormatterConsole FormatterType = "console"
)

// Formatter is a simple interface which takes bytes and produces formatted output:
type Formatter interface {
	Output(output interface{}) error
}
