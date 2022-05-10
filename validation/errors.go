package validation

import (
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
)

type ValidationError struct {
	Message      string         `json:"message,omitempty"`
	ShortMessage string         `json:"short_message,omitempty"`
	Hint         string         `json:"hint,omitempty"`
	Pos          lexer.Position `json:"pos,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s - on line: %v", e.Message, e.Pos.Line)
}

func (e *ValidationError) Contents() ValidationError {
	return *e
}
