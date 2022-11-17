package codegenerator

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

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

type SourceCodeType string

const (
	SourceCodeTypeDefinition SourceCodeType = "definition"
	SourceCodeTypeJavaScript SourceCodeType = "javascript"
)

type SourceCode struct {
	Contents string
	Path     string
	Type     SourceCodeType
}

type GeneratedFile = SourceCode

// GenerateSDK will generate a fresh @teamkeel/sdk package into the node_modules
// of the target directory
func (g *Generator) GenerateSDK() ([]*GeneratedFile, error) {
	sourceCodes := []*SourceCode{}

	sourceCodes = append(sourceCodes, &SourceCode{
		Path:     "index.js",
		Type:     SourceCodeTypeJavaScript,
		Contents: g.sdkSrcCode(),
	})

	sourceCodes = append(sourceCodes, &SourceCode{
		Path:     "index.d.ts",
		Type:     SourceCodeTypeDefinition,
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
		Type:     SourceCodeTypeJavaScript,
		Contents: g.testingSrcCode(),
	})

	sourceCodes = append(sourceCodes, &SourceCode{
		Path:     "index.d.ts",
		Type:     SourceCodeTypeDefinition,
		Contents: g.testingTypeDefinitions(),
	})

	err := g.makeNpmPackage(TestingPackageName, sourceCodes)

	if err != nil {
		return nil, err
	}

	return sourceCodes, nil
}

func (g *Generator) testingSrcCode() (r string) {
	return r
}

func (g *Generator) testingTypeDefinitions() (r string) {
	return r
}

func (g *Generator) sdkSrcCode() (r string) {
	r += "const doSomething = () => 'hello';\n"
	r += "const variableName = '';\n"

	return r
}

func (g *Generator) sdkTypeDefinitions() (r string) {
	return r
}

//go:embed package.json.tmpl
var templatePackageJson string

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
		return TSTypeString
	case proto.Type_TYPE_MODEL:
		if t.Repeated {
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
		return TSTypeUnknown
	}
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
