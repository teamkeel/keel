package codegenerator

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

// This package is responsible for performing code generation of the dynamic @teamkeel/sdk + @teamkeel/testing
// node packages, which include dynamically generated javascript code as well as typescript definition files (d.ts)
// representing all of the relevant constructs (see 'relevant constructs' below) in a .keel schema file.

// Relevant constructs in a .keel schema and their outputs:
// - Model & field definitions => representative interface generated in index.d.ts
// - Enum definitions => enum will be generated in index.d.ts
// - Inputs => interface type representing input type will be generated in index.d.ts
// - Model definition => Model API class for interacting with database will be generated (js + d.ts)
// - Custom functions => javascript + index.d.ts wrapper code

type Generator struct {
	schema *proto.Schema
	dir    string
}

func NewGenerator(schema *proto.Schema, dir string) *Generator {
	return &Generator{
		schema: schema,
		dir:    dir,
	}
}

type SourceCode struct {
	Contents string
	Path     string
}

type GeneratedFile = SourceCode

// GenerateSDK will generate a fresh @teamkeel/sdk package into the node_modules
// of the target directory
func (g *Generator) GenerateSDK() ([]*GeneratedFile, error) {
	sourceCodes := []*SourceCode{}

	sourceCodes = append(sourceCodes, &SourceCode{
		Path:     "index.js",
		Contents: g.sdkSrcCode(),
	})

	sourceCodes = append(sourceCodes, &SourceCode{
		Path:     "index.d.ts",
		Contents: g.sdkTypeDefinitions(),
	})

	err := g.makeNpmPackage(SdkPackageName, sourceCodes)

	if err != nil {
		return nil, err
	}

	return sourceCodes, nil
}

// GenerateTesting will generate a fresh @teamkeel/testing package into the node_modules
// of the target directory
func (g *Generator) GenerateTesting() ([]*GeneratedFile, error) {
	sourceCodes := []*SourceCode{}

	sourceCodes = append(sourceCodes, &SourceCode{
		Path:     "index.js",
		Contents: g.testingSrcCode(),
	})

	sourceCodes = append(sourceCodes, &SourceCode{
		Path:     "index.d.ts",
		Contents: g.testingTypeDefinitions(),
	})

	err := g.makeNpmPackage(TestingPackageName, sourceCodes)

	if err != nil {
		return nil, err
	}

	return sourceCodes, nil
}

func (g *Generator) GenerateDevelopmentHandler() ([]*GeneratedFile, error) {
	src := renderTemplate(TemplateHandlerDevelopment, map[string]interface{}{})

	sourceCodes := []*SourceCode{}

	sourceCodes = append(sourceCodes, &SourceCode{Path: "index.js", Contents: src})

	err := g.makeBuildDir(sourceCodes)

	if err != nil {
		return nil, err
	}

	return sourceCodes, nil
}

// Generates the contents of the index.js file, containing vanilla
// javascript code required by the testing package
func (g *Generator) testingSrcCode() string {
	return renderTemplate(TemplateTesting, map[string]interface{}{
		"Actions": g.schemaActions(),
		"Models":  g.schemaModels(),
	})
}

// Generates the contents of the index.d.ts file, containing typescript
// type definitions for the index.js javascript file.
func (g *Generator) testingTypeDefinitions() string {
	return renderTemplate(TemplateTestingDefinitions, map[string]interface{}{
		"Actions": g.schemaActions(),
		"Models":  g.schemaModels(),
	})
}

// Generates the contents of the index.js file, containing vanilla javascript code required by the sdk
func (g *Generator) sdkSrcCode() string {
	return renderTemplate(TemplateSdk, map[string]interface{}{
		"Models":  g.schemaModels(),
		"Actions": g.schemaActions(),
	})
}

// Generates the contents of the index.d.ts file, containing typescript
// type definitions for the index.js javascript file.
func (g *Generator) sdkTypeDefinitions() string {
	return renderTemplate(TemplateSdkDefinitions, map[string]interface{}{
		"Models":  g.schemaModels(),
		"Enums":   g.schemaEnums(),
		"Actions": g.schemaActions(),
	})
}

//go:embed package.json.tmpl
var templatePackageJson string

const (
	BUILD_DIR_NAME = ".build"
)

// makeBuildDir will create a hidden .build directory in the target directory,
// containing the given source code files
func (g *Generator) makeBuildDir(srcCodes []*SourceCode) error {
	basePath := filepath.Join(g.dir, BUILD_DIR_NAME)

	err := os.MkdirAll(basePath, os.ModePerm)

	if err != nil {
		return err
	}

	for _, code := range srcCodes {
		f, err := os.Create(filepath.Join(basePath, code.Path))

		if err != nil {
			return err
		}

		_, err = f.WriteString(code.Contents)

		if err != nil {
			return err
		}
	}

	return nil
}

// makeNpmPackage will create a new node_module at {dir}/node_modules/@teamkeel/{name},
// with a simple package.json, as well as any srcCode files passed to this method.
func (g *Generator) makeNpmPackage(name string, srcCodes []*SourceCode) error {
	basePath := filepath.Join(g.dir, "node_modules", "@teamkeel", name)
	err := os.MkdirAll(filepath.Join(basePath), os.ModePerm)

	if err != nil {
		return err
	}

	template, err := template.New("").Parse(templatePackageJson)

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	templateVars := map[string]any{
		"Name": name,
	}

	err = template.Execute(&buf, templateVars)

	if err != nil {
		return err
	}

	packageJsonPath := filepath.Join(basePath, "package.json")
	f, err := os.Create(packageJsonPath)

	if err != nil {
		return err
	}

	_, err = f.WriteString(buf.String())

	if err != nil {
		return err
	}

	for _, code := range srcCodes {
		f, err := os.Create(filepath.Join(basePath, code.Path))

		if err != nil {
			return err
		}

		_, err = f.WriteString(code.Contents)

		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) schemaEnums() (enums []*Enum) {
	for _, enum := range g.schema.Enums {
		e := Enum{
			Name: enum.Name,
		}

		for _, v := range enum.Values {
			e.Values = append(e.Values, &EnumValue{
				Label: v.Name,
			})
		}

		enums = append(enums, &e)
	}

	return enums
}

func (g *Generator) schemaModels() (models []*Model) {
	for _, model := range g.schema.Models {
		m := Model{
			Name:           model.Name,
			TableName:      strcase.ToSnake(model.Name),
			ApiName:        fmt.Sprintf("%sApi", model.Name),
			NameLowerCamel: strcase.ToLowerCamel(model.Name),
		}

		for _, field := range model.Fields {
			mf := &ModelField{
				Name:           field.Name,
				Type:           protoTypeToTypeScriptType(field.Type),
				ConstraintType: constraintTypeForField(field.Type),
				IsOptional:     field.Optional,
			}

			m.Fields = append(m.Fields, mf)

			if field.Unique || field.PrimaryKey {
				m.UniqueFields = append(m.UniqueFields, mf)
			}
		}

		models = append(models, &m)
	}

	return models
}

func (g *Generator) schemaActions() []*Action {
	return lo.Map(proto.FilterOperations(g.schema, func(op *proto.Operation) bool {
		return true
	}), func(op *proto.Operation, _ int) *Action {
		writeInputs := lo.Filter(op.Inputs, func(i *proto.OperationInput, _ int) bool {
			return i.Mode == proto.InputMode_INPUT_MODE_WRITE
		})

		readInputs := lo.Filter(op.Inputs, func(i *proto.OperationInput, _ int) bool {
			return i.Mode == proto.InputMode_INPUT_MODE_READ
		})

		return &Action{
			Name:           strcase.ToCamel(op.Name),
			OperationType:  operationTypeForOperation(op),
			NameLowerCamel: strcase.ToLowerCamel(op.Name),
			ModelName:      op.ModelName,
			IsCustom:       op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM,
			WriteInputs: lo.Map(writeInputs, func(i *proto.OperationInput, _ int) *ActionInput {
				return protoInputToActionInput(i)
			}),
			ReadInputs: lo.Map(readInputs, func(i *proto.OperationInput, _ int) *ActionInput {
				return protoInputToActionInput(i)
			}),
			// Some operation types will need all of the inputs no matter the mode (including Unknown mode for authenticate actions)
			Inputs: lo.Map(op.Inputs, func(i *proto.OperationInput, _ int) *ActionInput {
				return protoInputToActionInput(i)
			}),
		}
	})
}

//go:embed templates/*.tmpl
var templates embed.FS

type Typer interface {
	GetType() proto.Type
	IsRepeated() bool
}

// Maps the internal Keel field type to a corresponding valid typescript type
func protoTypeToTypeScriptType(t *proto.TypeInfo) string {
	switch t.GetType() {
	case proto.Type_TYPE_UNKNOWN:
		return TSTypeUnknown
	case proto.Type_TYPE_STRING:
		return TSTypeString
	case proto.Type_TYPE_BOOL:
		return TSTypeBoolean
	case proto.Type_TYPE_INT:
		return TSTypeNumber
	case proto.Type_TYPE_TIMESTAMP:
		return TSTypeTimestamp
	case proto.Type_TYPE_DATE:
		return TSTypeDate
	case proto.Type_TYPE_ID:
		return TSTypeID
	case proto.Type_TYPE_MODEL:
		if t.Repeated {
			// e.g Post[]
			return fmt.Sprintf("%s[]", t.ModelName.Value)
		}
		return t.ModelName.Value
	// case proto.Type_TYPE_CURRENCY:
	// 	return "Currency"
	case proto.Type_TYPE_DATETIME:
		return TSTypeDate
	case proto.Type_TYPE_ENUM:
		return t.EnumName.Value
	// case proto.Type_TYPE_IMAGE:
	// 	return "Image"
	case proto.Type_TYPE_PASSWORD: // todo: remove this and hide password fields going forward?
		return TSTypeString
	default:
		// for anything else, utilize typescript's unknown type
		// so that end user must be careful with the contents of the var:
		// https://stackoverflow.com/a/51441168/1795862
		return TSTypeUnknown
	}
}

func constraintTypeForField(t *proto.TypeInfo) string {
	if t.Type == proto.Type_TYPE_ENUM {
		return "EnumConstraint"
	}

	return fmt.Sprintf("%sConstraint", strcase.ToCamel(protoTypeToTypeScriptType(t)))
}

func renderTemplate(name string, data map[string]interface{}) string {
	template, err := template.ParseFS(templates, fmt.Sprintf("templates/%s.tmpl", name))
	if err != nil {
		panic(err)
	}
	var tpl bytes.Buffer

	err = template.Execute(&tpl, data)

	if err != nil {
		panic(err)
	}

	return tpl.String()
}

func inputModeStringFromInputMode(inputMode proto.InputMode) InputMode {
	switch inputMode {
	case proto.InputMode_INPUT_MODE_READ:
		return InputModeRead
	case proto.InputMode_INPUT_MODE_WRITE:
		return InputModeWrite
	default:
		return InputModeUnknown
	}
}

func protoInputToActionInput(input *proto.OperationInput) *ActionInput {
	return &ActionInput{
		Label:          input.Name,
		Type:           protoTypeToTypeScriptType(input.Type),
		IsOptional:     input.Optional,
		ConstraintType: constraintTypeForField(input.Type),
		Mode:           inputModeStringFromInputMode(input.Mode),
	}
}

// Go templates do not have support for comparing against complex logics
// we could compare against the underlying proto.OperationType enum values but
// it would make the templates really ugly, so in the interest of code succinctness
// create a friendly api around the proto object
func operationTypeForOperation(op *proto.Operation) OperationType {
	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
		return OperationTypeAuthenticate
	case proto.OperationType_OPERATION_TYPE_CREATE:
		return OperationTypeCreate
	case proto.OperationType_OPERATION_TYPE_DELETE:
		return OperationTypeDelete
	case proto.OperationType_OPERATION_TYPE_LIST:
		return OperationTypeList
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		return OperationTypeUpdate
	case proto.OperationType_OPERATION_TYPE_GET:
		return OperationTypeGet
	default:
		panic("Unknown operation type")
	}
}
