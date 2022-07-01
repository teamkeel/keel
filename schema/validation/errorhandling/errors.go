package errorhandling

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/util/str"

	"gopkg.in/yaml.v3"
)

// error codes
const (
	ErrorUpperCamel                         = "E001"
	ErrorActionNameLowerCamel               = "E002"
	ErrorFieldNamesUniqueInModel            = "E003"
	ErrorOperationsUniqueGlobally           = "E004"
	ErrorInvalidActionInput                 = "E005"
	ErrorReservedFieldName                  = "E006"
	ErrorReservedModelName                  = "E007"
	ErrorOperationMissingUniqueInput        = "E008"
	ErrorUnsupportedFieldType               = "E009"
	ErrorUniqueModelsGlobally               = "E010"
	ErrorUnsupportedAttributeType           = "E011"
	ErrorFieldNameLowerCamel                = "E012"
	ErrorInvalidAttributeArgument           = "E013"
	ErrorAttributeRequiresNamedArguments    = "E014"
	ErrorAttributeMissingRequiredArgument   = "E015"
	ErrorInvalidValue                       = "E016"
	ErrorUniqueAPIGlobally                  = "E017"
	ErrorUniqueRoleGlobally                 = "E018"
	ErrorUniqueEnumGlobally                 = "E019"
	ErrorUnresolvableExpression             = "E020"
	ErrorUnresolvedRootModel                = "E021"
	ErrorForbiddenExpressionOperation       = "E022"
	ErrorForbiddenValueCondition            = "E023"
	ErrorTooManyArguments                   = "E024"
	ErrorInvalidSyntax                      = "E025"
	ErrorExpressionTypeMismatch             = "E026"
	ErrorForbiddenOperator                  = "E027"
	ErrorNonBooleanValueCondition           = "E028"
	ErrorExpressionArrayWrongType           = "E029"
	ErrorExpressionArrayMismatchingOperator = "E030"
	ErrorExpressionForbiddenArrayLHS        = "E031"
	ErrorExpressionMixedTypesInArrayLiteral = "E032"
	ErrorCreateOperationNoInputs            = "E033"
	ErrorCreateOperationMissingInput        = "E034"
	ErrorOperationInputNotUnique            = "E035"
	ErrorOperationWhereNotUnique            = "E036"
	ErrorNonDirectComparisonOperatorUsed    = "E037"
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
	*ErrorDetails

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

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s - on line: %v", e.Message, e.Pos.Line)
}

func (e *ValidationError) Unwrap() error { return e }

type ValidationErrors struct {
	Errors []*ValidationError
}

func (v ValidationErrors) Error() string {
	str := ""

	for _, err := range v.Errors {
		str += fmt.Sprintf("%s: %s\n", err.Code, err.Message)
	}

	return str
}

// Returns the console flavoured output format for a set of validation errors
func (v ValidationErrors) ToConsole(sources []reader.SchemaFile) (string, error) {
	errorCount := len(v.Errors)
	errorsPartial := ""
	if errorCount > 1 {
		errorsPartial = "errors"
	} else {
		errorsPartial = "error"
	}

	statusMessage := red.Sprint("INVALID\n")
	errorCountMessage := yellow.Sprintf("%d validation %s:", len(v.Errors), errorsPartial)

	schemaPreview, err := v.ToAnnotatedSchema(sources)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n%s\n%s", statusMessage, errorCountMessage, schemaPreview), nil
}

// Returns a visual representation of a schema file, annotated with error highlighting and messages
func (v ValidationErrors) ToAnnotatedSchema(sources []reader.SchemaFile) (string, error) {
	schemaString := ""

	bufferLines := 5
	gutterAmount := 5
	newLine := func() {
		schemaString += "\n"
	}

	for _, err := range v.Errors {
		errorStartLine := err.Pos.Line
		errorEndLine := err.EndPos.Line

		var source string
		for _, s := range sources {
			if s.FileName == err.Pos.Filename {
				source = s.Contents
				break
			}
		}

		// kind of feels like this should be an error...
		if source == "" {
			return "", fmt.Errorf("no source file provided for %s", err.Pos.Filename)
		}

		lines := strings.Split(source, "\n")
		codeStartCol := len(fmt.Sprintf("%d", len(lines))) + gutterAmount
		midPointPosition := codeStartCol + err.Pos.Column + ((err.EndPos.Column - err.Pos.Column) / 2)
		tokenLength := err.EndPos.Column - err.Pos.Column

		for lineIndex, line := range lines {
			// Render line numbers in gutter
			outputLine := blue.Sprint(str.PadRight(fmt.Sprintf("%d", lineIndex+1), codeStartCol))

			// If this line isn't close enough to an error let's ignore it
			if (lineIndex+1) < (errorStartLine-bufferLines) || (lineIndex+1) > (errorEndLine+bufferLines) {
				continue
			}

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

			// Begin closures to render unicode arrows / hints / messages
			indent := func(length int) {
				counter := 1

				for counter < length {
					schemaString += " "
					counter += 1
				}
			}

			underline := func() {
				indent(codeStartCol + err.Pos.Column)

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

			arrowDown := func() {
				newLine()
				indent(midPointPosition)
				schemaString += yellow.Sprint("\u2570")
				schemaString += yellow.Sprint("\u2500")
			}

			message := func() {
				schemaString += fmt.Sprintf(" %s %s", yellow.Sprint(err.ErrorDetails.Message), red.Sprintf("(%s)", err.Code))
			}

			hint := func() {
				if err.ErrorDetails.Hint != "" {
					schemaString += cyan.Sprint(err.ErrorDetails.Hint)
				}
			}

			underline()
			arrowDown()
			message()
			newLine()

			// Line up hint with the error message above (taking into account unicode arrows)
			hintOffset := 3
			indent(midPointPosition + hintOffset)
			hint()
			newLine()
		}

		schemaString += red.Add(color.Italic).Sprintf("\u21B3 %s", err.Pos.Filename)
		newLine()
		newLine()

	}

	return schemaString, nil
}

func (e ValidationErrors) Unwrap() error { return e }

func NewValidationError(code string, data TemplateLiterals, position node.ParserNode) *ValidationError {
	start, end := position.GetPositionRange()

	return &ValidationError{
		Code: code,
		// todo global locale setting
		ErrorDetails: buildErrorDetailsFromYaml(code, "en", data),
		Pos: LexerPos{
			Filename: start.Filename,
			Offset:   start.Offset,
			Line:     start.Line,
			Column:   start.Column,
		},
		EndPos: LexerPos{
			Filename: end.Filename,
			Offset:   end.Offset,
			Line:     end.Line,
			Column:   end.Column,
		},
	}
}

//go:embed errors.yml
var errorsYaml []byte

var errorDetailsByCode map[string]map[string]*ErrorDetails

func init() {
	err := yaml.Unmarshal(errorsYaml, &errorDetailsByCode)

	if err != nil {
		panic(err)
	}
}

func renderTemplate(tmpl string, data map[string]string) string {
	template, err := template.New("template").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = template.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

// Takes an error code like E001, finds the relevant copy in the errors.yml file and interpolates the literals into the yaml template.
func buildErrorDetailsFromYaml(code string, locale string, literals TemplateLiterals) *ErrorDetails {
	errorDetails, ok := errorDetailsByCode[locale][code]
	if !ok {
		panic(fmt.Sprintf("no error details for error code: %s", code))
	}

	return &ErrorDetails{
		Message:      renderTemplate(errorDetails.Message, literals.Literals),
		ShortMessage: renderTemplate(errorDetails.ShortMessage, literals.Literals),
		Hint:         renderTemplate(errorDetails.Hint, literals.Literals),
	}
}
