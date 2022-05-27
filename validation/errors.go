package validation

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/davecgh/go-spew/spew"
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
	Message      string `json:"message,omitempty" yaml:"message"`
	ShortMessage string `json:"short_message,omitempty" yaml:"short_message"`
	Hint         string `json:"hint,omitempty" yaml:"hint"`
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
	ret := ""

	red := color.New(color.FgRed)
	matchingSchemas := v.MatchingSchemas()

	for _, err := range v.Errors {
		errorStartLine := err.Pos.Line
		errorEndLine := err.EndPos.Line
		errorStartColumn := err.Pos.Column
		errorEndColumn := err.EndPos.Column
		spew.Dump("start:", err.Pos)
		spew.Dump("end:", err.EndPos)

		if match, ok := matchingSchemas[err.Pos.Filename]; ok {
			lines := strings.Split(match.Contents, "\n")

			for lineIndex, line := range lines {
				if (lineIndex+1) < errorStartLine || (lineIndex+1) > errorEndLine {
					ret += fmt.Sprintf("%s\n", line)

					continue
				}

				outputLine := ""
				chars := strings.Split(line, "")

				for charIdx, char := range chars {

					if (charIdx+1) < errorStartColumn && (charIdx+1) < errorEndColumn {
						outputLine += char
						continue
					}

					outputLine += red.Sprint(char)
				}

				ret += fmt.Sprintf("%s\n", outputLine)

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

func (v ValidationErrors) AsBytes() []byte {

	return []byte(
		v.Error(),
	)
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
