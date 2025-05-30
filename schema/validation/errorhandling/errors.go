package errorhandling

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/teamkeel/keel/schema/node"

	"gopkg.in/yaml.v3"
)

// error codes
const (
	ErrorUpperCamel                      = "E001"
	ErrorActionNameLowerCamel            = "E002"
	ErrorFieldNamesUniqueInModel         = "E003"
	ErrorActionUniqueGlobally            = "E004"
	ErrorInvalidActionInput              = "E005"
	ErrorReservedFieldName               = "E006"
	ErrorUnsupportedFieldType            = "E009"
	ErrorUniqueModelsGlobally            = "E010"
	ErrorUnsupportedAttributeType        = "E011"
	ErrorInvalidAttributeArgument        = "E013"
	ErrorAttributeRequiresNamedArguments = "E014"
	ErrorUniqueAPIGlobally               = "E017"
	ErrorUniqueRoleGlobally              = "E018"
	ErrorForbiddenExpressionAction       = "E022"
	ErrorInvalidSyntax                   = "E025"
	ErrorCreateActionNoInputs            = "E033"
	ErrorCreateActionMissingInput        = "E034"
	ErrorInvalidActionType               = "E040"
	ErrorModelNotFound                   = "E047"
	ErrorFieldNamesMaxLength             = "E052"
	ErrorModelNamesMaxLength             = "E053"
)

type ErrorDetails struct {
	Message string `json:"message" yaml:"message"`
	Hint    string `json:"hint"    yaml:"hint"`
}

type TemplateLiterals struct {
	Literals map[string]string
}

type ValidationError struct {
	*ErrorDetails

	Code   string   `json:"code"             regexp:"\\d+"`
	Pos    LexerPos `json:"pos,omitempty"`
	EndPos LexerPos `json:"endPos,omitempty"`
}

type LexerPos struct {
	Filename string `json:"filename"`
	Offset   int    `json:"offset"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s - on line: %v", e.Message, e.Pos.Line)
}

func (e *ValidationError) Unwrap() error { return e }

type ValidationErrors struct {
	Errors   []*ValidationError `json:"errors"`
	Warnings []*ValidationError `json:"warnings"`
}

func (v *ValidationErrors) Append(code string, data map[string]string, node node.ParserNode) {
	v.Errors = append(v.Errors, NewValidationError(code,
		TemplateLiterals{
			Literals: data,
		},
		node,
	))
}

func (v *ValidationErrors) AppendWarning(e *ValidationError) {
	if e != nil {
		v.Warnings = append(v.Warnings, e)
	}
}

func (v *ValidationErrors) AppendError(e *ValidationError) {
	if e != nil {
		v.Errors = append(v.Errors, e)
	}
}

func (v *ValidationErrors) AppendErrors(errs []*ValidationError) {
	v.Errors = append(v.Errors, errs...)
}

func (v *ValidationErrors) Concat(verrs ValidationErrors) {
	v.Errors = append(v.Errors, verrs.Errors...)
}

func (v ValidationErrors) Error() string {
	str := ""
	for _, err := range v.Errors {
		str += fmt.Sprintf("%s: %s\n", err.Code, err.Message)
	}

	return str
}

func (v ValidationErrors) Warning() string {
	str := ""
	for _, err := range v.Warnings {
		str += fmt.Sprintf("%s: %s\n", err.Code, err.Message)
	}

	return str
}

func (v ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v ValidationErrors) HasWarnings() bool {
	return len(v.Warnings) > 0
}

type ErrorType string

const (
	NamingError              ErrorType = "NamingError"
	DuplicateDefinitionError ErrorType = "DuplicateDefinitionError"
	TypeError                ErrorType = "TypeError"
	UndefinedError           ErrorType = "UndefinedError"
	ActionInputError         ErrorType = "ActionInputError"
	AttributeArgumentError   ErrorType = "AttributeArgumentError"
	AttributeNotAllowedError ErrorType = "AttributeNotAllowedError"
	AttributeExpressionError ErrorType = "AttributeExpressionError"
	RelationshipError        ErrorType = "RelationshipError"
	JobDefinitionError       ErrorType = "JobDefinitionError"
	UnsupportedFeatureError  ErrorType = "UnsupportedFeatureError"
)

func NewValidationErrorWithDetails(t ErrorType, details ErrorDetails, position node.ParserNode) *ValidationError {
	start, end := position.GetPositionRange()

	return &ValidationError{
		Code:         string(t),
		ErrorDetails: &details,
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

func renderTemplate(name string, tmpl string, data map[string]string) string {
	template, err := template.New(name).Parse(tmpl)
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
	ed, ok := errorDetailsByCode[locale][code]
	if !ok {
		panic(fmt.Sprintf("no error details for error code: %s", code))
	}

	return &ErrorDetails{
		Message: renderTemplate(fmt.Sprintf("%s-%s", code, "message"), ed.Message, literals.Literals),
		Hint:    renderTemplate(fmt.Sprintf("%s-%s", code, "hint"), ed.Hint, literals.Literals),
	}
}
