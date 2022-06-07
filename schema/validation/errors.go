package validation

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/fatih/color"
	"github.com/teamkeel/keel/model"

	"gopkg.in/yaml.v3"
)

// error codes
const (
	ErrorUpperCamel                   = "E001"
	ErrorOperationNameLowerCamel      = "E002"
	ErrorFieldNamesUniqueInModel      = "E003"
	ErrorOperationsUniqueGlobally     = "E004"
	ErrorInputsNotFields              = "E005"
	ErrorReservedFieldName            = "E006"
	ErrorReservedModelName            = "E007"
	ErrorOperationInputFieldNotUnique = "E008"
	ErrorUnsupportedFieldType         = "E009"
	ErrorUniqueModelsGlobally         = "E010"
	ErrorUnsupportedAttributeType     = "E011"
	ErrorFieldNameLowerCamel          = "E012"
	ErrorFunctionNameLowerCamel       = "E013"
)

type ErrorDetails struct {
	Message      string `json:"message" yaml:"message"`
	ShortMessage string `json:"short_message" yaml:"short_message"`
	Hint         string `json:"hint" yaml:"hint"`
}

type TemplateLiterals struct {
	Literals map[string]string
}

type ValidationError struct {
	ErrorDetails

	Code   string   `json:"code" regexp:"\\d+"`
	Pos    LexerPos `json:"pos,omitempty"`
	EndPos LexerPos `json:"end_pos,omitempty"`
}

type LexerPos struct {
	Filename string `json:"filename"`
	Offset   int    `json:"offset"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

var red, blue, yellow, cyan color.Color = *color.New(color.FgRed), *color.New(color.FgHiBlue), *color.New(color.FgHiYellow), *color.New(color.FgCyan)

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s - on line: %v", e.Message, e.Pos.Line)
}

func (e *ValidationError) Unwrap() error { return e }

type ValidationErrors struct {
	Errors []*ValidationError
}

func (v ValidationErrors) MatchingSchemas() map[string]model.SchemaFile {
	paths := []string{}
	schemaFiles := map[string]model.SchemaFile{}

	for _, err := range v.Errors {
		if contains(paths, err.Pos.Filename) {
			continue
		}

		paths = append(paths, err.Pos.Filename)
	}

	for _, path := range paths {
		fileBytes, err := os.ReadFile(path)

		if err != nil {
			panic(err)
		}

		schemaFiles[path] = model.SchemaFile{FileName: path, Contents: string(fileBytes)}
	}

	return schemaFiles
}

func (v ValidationErrors) Error() string {
	str := ""

	for _, err := range v.Errors {
		str += fmt.Sprintf("%s\n", err.Message)
	}

	return str
}

// Returns the console flavoured output format for a set of validation errors
func (v ValidationErrors) ToConsole() string {
	errorCount := len(v.Errors)
	errorsPartial := ""
	if errorCount > 1 {
		errorsPartial = "errors"
	} else {
		errorsPartial = "error"
	}

	statusMessage := red.Sprint("INVALID\n")
	errorCountMessage := yellow.Sprintf("%d validation %s:", len(v.Errors), errorsPartial)

	schemaPreview := v.ToAnnotatedSchema()

	return fmt.Sprintf("%s\n%s\n%s", statusMessage, errorCountMessage, schemaPreview)
}

// Returns a visual representation of a schema file, annotated with error highlighting and messages
func (v ValidationErrors) ToAnnotatedSchema() string {
	schemaString := ""

	matchingSchemas := v.MatchingSchemas()

	gutterAmount := 5

	for _, err := range v.Errors {
		errorStartLine := err.Pos.Line
		errorEndLine := err.EndPos.Line

		if match, ok := matchingSchemas[err.Pos.Filename]; ok {
			lines := strings.Split(match.Contents, "\n")
			codeStartCol := len(fmt.Sprintf("%d", len(lines))) + gutterAmount

			for lineIndex, line := range lines {
				// Render line numbers in gutter
				outputLine := blue.Sprint(padRight(fmt.Sprintf("%d", lineIndex+1), codeStartCol))

				// If the error line doesn't match the currently enumerated line
				// then we can render the whole line without any colorization
				if (lineIndex+1) < errorStartLine || (lineIndex+1) > errorEndLine {
					outputLine += fmt.Sprintf("%s\n", line)

					schemaString += outputLine
					continue
				}

				chars := strings.Split(line, "")

				// Enumerate over the characters in the line
				for charIdx, char := range chars {

					// Check if the character index is less than or greater than the corresponding start and end column
					// If so, then render the char without any colorization
					if (charIdx+1) < err.Pos.Column || (charIdx+1) > err.EndPos.Column-1 {
						outputLine += char
						continue
					}

					outputLine += red.Sprint(char)
				}

				schemaString += fmt.Sprintf("%s\n", outputLine)

				// Find the token in the char array based on the start and end column position of the error
				token := strings.TrimSpace(strings.Join(chars[err.Pos.Column-1:err.EndPos.Column-1], ""))

				// Find the midpoint of the token in the wider context of the line
				// The codeStartCol is the the sum of the number of digits of the rendered line number + the default gutter of 5
				midPointPosition := codeStartCol + err.Pos.Column + (len(token) / 2)

				// Begin closures to render unicode arrows / hints / messages
				newLine := func() {
					schemaString += "\n"
				}

				indent := func(length int) {
					counter := 1

					for counter < length {
						schemaString += " "
						counter += 1
					}
				}

				underline := func(token string) {
					indent(codeStartCol + err.Pos.Column)

					tokenLength := len(token)

					counter := 0

					for counter < tokenLength {
						if counter == tokenLength/2 {
							schemaString += yellow.Sprint("\u252C")
						} else {
							schemaString += yellow.Sprint("\u2500")

						}
						counter++
					}
				}

				arrowDown := func(token string) {
					newLine()
					indent(midPointPosition)
					schemaString += yellow.Sprint("\u2570")
					schemaString += yellow.Sprint("\u2500")
				}

				message := func() {
					schemaString += yellow.Sprintf(" %s", err.ErrorDetails.Message)
				}

				hint := func() {
					schemaString += cyan.Sprint(err.ErrorDetails.Hint)
				}

				underline(token)
				arrowDown(token)
				message()
				newLine()

				// Line up hint with the error message above (taking into account unicode arrows)
				hintOffset := 3
				indent(midPointPosition + hintOffset)
				hint()
				newLine()
			}
		}
	}

	return schemaString
}

func (e ValidationErrors) Unwrap() error { return e }

func validationError(code string, data TemplateLiterals, Pos lexer.Position, EndPos lexer.Position) error {
	return &ValidationError{
		Code: code,
		// todo global locale setting
		ErrorDetails: *buildErrorDetailsFromYaml(code, "en", data),
		Pos: LexerPos{
			Filename: Pos.Filename,
			Offset:   Pos.Offset,
			Line:     Pos.Line,
			Column:   Pos.Column,
		},
		EndPos: LexerPos{
			Filename: EndPos.Filename,
			Offset:   EndPos.Offset,
			Line:     EndPos.Line,
			Column:   EndPos.Column,
		},
	}
}

//go:embed errors.yml
var fileBytes []byte

// Takes an error code like E001, finds the relevant copy in the errors.yml file and interpolates the literals into the yaml template.
func buildErrorDetailsFromYaml(code string, locale string, literals TemplateLiterals) *ErrorDetails {
	m := make(map[string]map[string]interface{})

	err := yaml.Unmarshal(fileBytes, &m)

	if err != nil {
		panic(err)
	}

	slice := m[locale][code]

	sliceYaml, err := yaml.Marshal(slice)

	if err != nil {
		panic(err)
	}

	template, err := template.New(code).Parse(string(sliceYaml))

	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = template.Execute(&buf, literals.Literals)

	if err != nil {
		panic(err)
	}

	interpolatedBytes := buf.Bytes()

	o := make(map[string]string)

	err = yaml.Unmarshal(interpolatedBytes, &o)

	if err != nil {
		panic(err)
	}

	return &ErrorDetails{
		Message:      o["message"],
		ShortMessage: o["short_message"],
		Hint:         o["hint"],
	}
}

func padRight(str string, padAmount int) string {
	for len(str) < padAmount {
		str += " "
	}

	return str
}
