package validation

import (
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
)

type ValidationError struct {
	Message      string   `json:"message,omitempty"`
	ShortMessage string   `json:"short_message,omitempty"`
	Hint         string   `json:"hint,omitempty"`
	Pos          LexerPos `json:"pos,omitempty"`
}

type LexerPos struct {
	Filename string `json:"filename,omitempty"`
	Offset   int    `json:"offset,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s - on line: %v", e.Message, e.Pos.Line)
}

func (e *ValidationError) Unwrap() error { return e }

type ValidationErrors struct {
	Errors []*ValidationError
}

func (v ValidationErrors) Error() string {
	return fmt.Sprintf("%d validation errors found", len(v.Errors))
}

func (e ValidationErrors) Unwrap() error { return e }

func validationError(message, shortMessage, hint string, Pos lexer.Position) error {
	return &ValidationError{
		Message:      message,
		ShortMessage: shortMessage,
		Hint:         hint,
		Pos: LexerPos{
			Filename: Pos.Filename,
			Offset:   Pos.Offset,
			Line:     Pos.Line,
			Column:   Pos.Column,
		},
	}
}
