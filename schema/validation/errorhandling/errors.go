package errorhandling

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/teamkeel/keel/schema/associations"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/util/collection"
	"github.com/teamkeel/keel/util/str"

	"gopkg.in/yaml.v3"
)

// error codes
const (
	ErrorUpperCamel                       = "E001"
	ErrorActionNameLowerCamel             = "E002"
	ErrorFieldNamesUniqueInModel          = "E003"
	ErrorOperationsUniqueGlobally         = "E004"
	ErrorInvalidActionInput               = "E005"
	ErrorReservedFieldName                = "E006"
	ErrorReservedModelName                = "E007"
	ErrorOperationInputFieldNotUnique     = "E008"
	ErrorUnsupportedFieldType             = "E009"
	ErrorUniqueModelsGlobally             = "E010"
	ErrorUnsupportedAttributeType         = "E011"
	ErrorFieldNameLowerCamel              = "E012"
	ErrorInvalidAttributeArgument         = "E013"
	ErrorAttributeRequiresNamedArguments  = "E014"
	ErrorAttributeMissingRequiredArgument = "E015"
	ErrorInvalidValue                     = "E016"
	ErrorUniqueAPIGlobally                = "E017"
	ErrorUniqueRoleGlobally               = "E018"
	ErrorUniqueEnumGlobally               = "E019"
	ErrorUnresolvableExpression           = "E020"
	ErrorUnresolvedRootModel              = "E021"
	ErrorForbiddenExpressionOperation     = "E022"
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
var bgWhite = *color.New(color.BgWhite)

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s - on line: %v", e.Message, e.Pos.Line)
}

func (e *ValidationError) Unwrap() error { return e }

type ValidationErrors struct {
	Errors []*ValidationError
}

func (v ValidationErrors) MatchingSchemas() map[string]reader.SchemaFile {
	paths := []string{}
	schemaFiles := map[string]reader.SchemaFile{}

	for _, err := range v.Errors {
		if collection.Contains(paths, err.Pos.Filename) {
			continue
		}

		paths = append(paths, err.Pos.Filename)
	}

	for _, path := range paths {
		fileBytes, err := os.ReadFile(path)

		if err != nil {
			panic(err)
		}

		schemaFiles[path] = reader.SchemaFile{FileName: path, Contents: string(fileBytes)}
	}

	return schemaFiles
}

func (v ValidationErrors) Error() string {
	str := ""

	for _, err := range v.Errors {
		str += fmt.Sprintf("%s: %s\n", err.Code, err.Message)
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
	newLine := func() {
		schemaString += "\n"
	}

	for _, err := range v.Errors {
		errorStartLine := err.Pos.Line
		errorEndLine := err.EndPos.Line

		if match, ok := matchingSchemas[err.Pos.Filename]; ok {
			lines := strings.Split(match.Contents, "\n")
			codeStartCol := len(fmt.Sprintf("%d", len(lines))) + gutterAmount
			midPointPosition := codeStartCol + err.Pos.Column + ((err.EndPos.Column - err.Pos.Column) / 2)
			tokenLength := err.EndPos.Column - err.Pos.Column

			for lineIndex, line := range lines {
				// Render line numbers in gutter
				outputLine := blue.Sprint(str.PadRight(fmt.Sprintf("%d", lineIndex+1), codeStartCol))

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
					schemaString += yellow.Sprintf(" %s", err.ErrorDetails.Message)
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
	}

	return schemaString
}

func (e ValidationErrors) Unwrap() error { return e }

func NewValidationError(code string, data TemplateLiterals, position node.ParserNode) error {
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

func NewAssociationValidationError(asts []*parser.AST, context interface{}, association *associations.Association) error {
	unresolved := association.UnresolvedFragment()
	suggestion := ""

	if len(association.Fragments) == 1 {
		// If there is only one fragment in the association
		// then it means that the root model was unresolvable
		// So therefore the suggestion should be the context (downcased)
		if model, ok := context.(*parser.ModelNode); ok {
			suggestion = strings.ToLower(model.Name.Value)
		}

		literals := map[string]string{
			"Type":  "association",
			"Root":  unresolved.Current,
			"Model": suggestion,
		}

		return NewValidationError(ErrorUnresolvedRootModel,
			TemplateLiterals{
				Literals: literals,
			},
			unresolved,
		)
	}

	// If more than one fragment, then we need to resolve the second fragment's parent
	// And find the field names on the parent in order to build up the suggestion hint
	// e.g Given the condition post.autho it should suggest post.author instead
	parentModel := query.Model(asts, unresolved.Parent)
	fieldsOnParent := query.ModelFieldNames(parentModel)

	correctionHint := NewCorrectionHint(fieldsOnParent, unresolved.Current)

	literals := map[string]string{
		"Type":       "association",
		"Fragment":   unresolved.Current,
		"Parent":     unresolved.Parent,
		"Suggestion": correctionHint.ToString(),
	}

	return NewValidationError(ErrorUnresolvableExpression,
		TemplateLiterals{
			Literals: literals,
		},
		unresolved.Node,
	)

}

//go:embed errors.yml
var fileBytes []byte

// Takes an error code like E001, finds the relevant copy in the errors.yml file and interpolates the literals into the yaml template.
func buildErrorDetailsFromYaml(code string, locale string, literals TemplateLiterals) ErrorDetails {
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

	return ErrorDetails{
		Message:      o["message"],
		ShortMessage: o["short_message"],
		Hint:         o["hint"],
	}
}
