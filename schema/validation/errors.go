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

func (v ValidationErrors) ToAnnotatedSchema() string {
	ret := ""

	red := color.New(color.FgRed)
	white := color.New(color.FgWhite).Add(color.FgHiBlue)

	matchingSchemas := v.MatchingSchemas()

	gutterAmount := 5

	for _, err := range v.Errors {
		errorStartLine := err.Pos.Line
		errorEndLine := err.EndPos.Line

		if match, ok := matchingSchemas[err.Pos.Filename]; ok {
			lines := strings.Split(match.Contents, "\n")
			codeStartCol := len(fmt.Sprintf("%d", len(lines))) + gutterAmount

			for lineIndex, line := range lines {
				// line numbers
				outputLine := white.Sprint(padRight(fmt.Sprintf("%d", lineIndex+1), codeStartCol))

				if (lineIndex+1) < errorStartLine || (lineIndex+1) > errorEndLine {
					outputLine += fmt.Sprintf("%s\n", line)

					ret += outputLine
					continue
				}

				chars := strings.Split(line, "")

				for charIdx, char := range chars {

					if (charIdx+1) < err.Pos.Column || (charIdx+1) > err.EndPos.Column-1 {
						outputLine += char
						continue
					}

					outputLine += red.Sprint(char)
				}

				ret += fmt.Sprintf("%s\n", outputLine)
				token := strings.TrimSpace(strings.Join(chars[err.Pos.Column-1:err.EndPos.Column-1], ""))
				midPointPosition := codeStartCol + err.Pos.Column + (len(token) / 2)

				newLine := func() {
					ret += "\n"
				}

				indent := func(length int) {
					counter := 1

					for counter < length {
						ret += " "
						counter += 1
					}
				}

				underline := func(token string) {
					indent(codeStartCol + err.Pos.Column)

					tokenLength := len(token)

					counter := 0

					for counter < tokenLength {
						if counter == tokenLength/2 {
							ret += color.New(color.FgYellow).Sprint("\u252C")
						} else {
							ret += color.New(color.FgYellow).Sprint("\u2500")

						}
						counter++
					}
				}

				arrowDown := func(token string) {
					newLine()
					indent(midPointPosition)
					ret += color.New(color.FgYellow).Sprint("\u2502")
					newLine()
					indent(midPointPosition)
					ret += color.New(color.FgYellow).Sprint("\u2570")
					ret += color.New(color.FgYellow).Sprint("\u2500")
				}

				underline(token)
				arrowDown(token)
				ret += fmt.Sprintf("%s\n", color.New(color.FgYellow).Sprintf(" %s. %s", err.ErrorDetails.Message, err.ErrorDetails.Hint))
				newLine()
			}
		}
	}

	errorCount := len(v.Errors)
	errorsPartial := ""
	if errorCount > 1 {
		errorsPartial = "errors"
	} else {
		errorsPartial = "error"
	}
	infoMessage := red.Add(color.Underline).Sprintf("%d validation %s found:", len(v.Errors), errorsPartial)
	schemaPreview := ret
	errorDetail := ""

	for _, err := range v.Errors {
		errorDetail += fmt.Sprintf("• %s \n", err.Message)
	}

	return fmt.Sprintf("%s\n%s\n%s", infoMessage, errorDetail, schemaPreview)
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

func MaxOf(ints ...int) int {
	max := ints[0]

	for _, i := range ints {
		if max < i {
			max = i
		}
	}

	return max
}

func MinOf(vars ...int) int {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}

func padRight(str string, padAmount int) string {
	for len(str) < padAmount {
		str += " "
	}

	return str
}

func recursionCountDigits(number int) int {
	if number < 10 {
		return 1
	} else {
		return 1 + recursionCountDigits(number/10)
	}
}
