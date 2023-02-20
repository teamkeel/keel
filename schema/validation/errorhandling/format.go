package errorhandling

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/teamkeel/keel/schema/reader"
)

var (
	red    = color.New(color.FgRed)
	green  = color.New(color.FgHiGreen, color.Faint)
	yellow = color.New(color.FgYellow)
	gray   = color.New(color.FgWhite, color.Faint)
)

// ToAnnotatedSchema formats the validation errors by pointing to the relevant line
// in the source file that produced the error
//
// The output is formatted using ANSI colours (if supported by the environment).
//
// To force colour on or off set the github.com/fatih/color.NoColor flag to true or
// false before calling this function.
func (verrs *ValidationErrors) ToAnnotatedSchema(sources []reader.SchemaFile) string {

	// Number of lines of the source code to render before and after the line with the error
	bufferLines := 3

	// How much to indent the entire result e.g. every line is indented this much
	indent := 2

	result := strings.Repeat(" ", indent)
	newLine := func() {
		result += "\n" + strings.Repeat(" ", indent)
	}

	for _, err := range verrs.Errors {
		// Assumption here is that the error is on one line
		errorLine := err.Pos.Line

		startSourceLine := errorLine - bufferLines
		endSourceLine := errorLine + bufferLines

		// This produces a format string like "%3s| " which we use to render the gutter
		// The number after the "%" is the width, which is documented as:
		//   > For most values, width is the minimum number of runes to output,
		//   > padding the formatted form with spaces if necessary.
		gutterFmt := "%" + fmt.Sprintf("%d", len(fmt.Sprintf("%d", endSourceLine))) + "s| "

		var source string
		for _, s := range sources {
			if s.FileName == err.Pos.Filename {
				source = s.Contents
				break
			}
		}

		result += green.Sprintf(gutterFmt, " ")
		result += green.Sprint(err.Pos.Filename)
		newLine()

		// not sure this can happen, but just in case we'll handle it
		if source == "" {
			result += err.Message
			newLine()
			continue
		}

		lines := strings.Split(source, "\n")

		for lineIndex, line := range lines {

			// If this line is outside of the buffer we can drop it
			if (lineIndex+1) < (startSourceLine) || (lineIndex+1) > (endSourceLine) {
				continue
			}

			// Render line numbers in gutter
			result += gray.Sprintf(gutterFmt, fmt.Sprintf("%d", lineIndex+1))

			// If the error line doesn't match the currently enumerated line
			// then we can render the whole line without any colorization
			if (lineIndex + 1) != errorLine {
				result += gray.Sprintf("%s", line)
				newLine()
				continue
			}

			chars := strings.Split(line, "")

			// Enumerate over the characters in the line
			for charIdx, char := range chars {

				// Check if the character index is less than or greater than the corresponding start and end column
				// If so, then render the char without any colorization
				if (charIdx+1) < err.Pos.Column || (charIdx+1) > err.EndPos.Column-1 {
					result += char
					continue
				}

				result += red.Sprint(char)
			}

			newLine()

			// Underline the token that caused the error
			result += gray.Sprintf(gutterFmt, "")
			result += strings.Repeat(" ", err.Pos.Column-1)
			tokenLength := err.EndPos.Column - err.Pos.Column
			for i := 0; i < tokenLength; i++ {
				if i == tokenLength/2 {
					result += yellow.Sprint("\u252C")
				} else {
					result += yellow.Sprint("\u2500")
				}
			}
			newLine()

			// Render the down arrow
			result += gray.Sprintf(gutterFmt, "")
			result += strings.Repeat(" ", err.Pos.Column-1)
			result += strings.Repeat(" ", (err.EndPos.Column-err.Pos.Column)/2)
			result += yellow.Sprint("\u2570")
			result += yellow.Sprint("\u2500")

			// Render the message
			result += fmt.Sprintf(" %s %s", yellow.Sprint(err.ErrorDetails.Message), red.Sprintf("(%s)", err.Code))
			newLine()

			// Render the hint
			if err.ErrorDetails.Hint != "" {
				result += gray.Sprintf(gutterFmt, "")
				result += strings.Repeat(" ", err.Pos.Column-1)
				result += strings.Repeat(" ", (err.EndPos.Column-err.Pos.Column)/2)
				// Line up hint with the error message above (taking into account unicode arrows)
				result += strings.Repeat(" ", 3)
				result += yellow.Sprint(err.ErrorDetails.Hint)
				newLine()
			}
		}

		newLine()
	}

	return result
}
