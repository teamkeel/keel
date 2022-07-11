package schema

import (
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/teamkeel/keel/proto"

	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/schema/validation"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// A Builder knows how to produce a (validated) proto.Schema,
// from a given Keel Builder. Construct one, then call the Make method.
type Builder struct {
	asts        []*parser.AST
	schemaFiles []reader.SchemaFile
}

// MakeFromDirectory constructs a proto.Schema from the .keel files present in the given
// directory.
func (scm *Builder) MakeFromDirectory(directory string) (*proto.Schema, error) {
	allInputFiles, err := reader.FromDir(directory)
	if err != nil {
		return nil, fmt.Errorf("error assembling input files: %v", err)
	}

	scm.schemaFiles = allInputFiles.SchemaFiles
	return scm.makeFromInputs(allInputFiles)
}

// MakeFromFile constructs a proto.Schema from the given .keel file.
func (scm *Builder) MakeFromFile(filename string) (*proto.Schema, error) {
	allInputFiles, err := reader.FromFile(filename)
	if err != nil {
		return nil, err
	}

	scm.schemaFiles = allInputFiles.SchemaFiles
	return scm.makeFromInputs(allInputFiles)
}

// MakeFromFile constructs a proto.Schema from the given inputs
func (scm *Builder) MakeFromInputs(inputs *reader.Inputs) (*proto.Schema, error) {
	scm.schemaFiles = inputs.SchemaFiles
	return scm.makeFromInputs(inputs)
}

func (scm *Builder) SchemaFiles() []reader.SchemaFile {
	return scm.schemaFiles
}

func (scm *Builder) ASTs() []*parser.AST {
	return scm.asts
}

func (scm *Builder) makeFromInputs(allInputFiles *reader.Inputs) (*proto.Schema, error) {
	// - For each of the .keel (schema) files specified...
	// 		- Parse to AST
	// 		- Add built-in fields
	// - With the parsed (AST) schemas as a set:
	// 		- Validate them (as a set)
	// 		- Convert the set to a single / aggregate proto model
	asts := []*parser.AST{}
	parseErrors := errorhandling.ValidationErrors{}
	for i, oneInputSchemaFile := range allInputFiles.SchemaFiles {
		declarations, err := parser.Parse(&oneInputSchemaFile)
		if err != nil {

			// try to convert into a validation error and move to next schema file
			if perr, ok := err.(parser.Error); ok {
				verr := errorhandling.NewValidationError(errorhandling.ErrorInvalidSyntax, errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Message": perr.Error(),
					},
				}, perr)
				parseErrors.Errors = append(parseErrors.Errors, verr)
				continue
			}

			return nil, fmt.Errorf("parser.Parse() failed on file: %s, with error %v", oneInputSchemaFile.FileName, err)
		}

		// Insert built in models like Identity. We only want to call this once
		// so that only one instance of the built in models are added if there
		// are multiple ASTs at play.
		// We want the insertion of built in models to happen
		// before insertion of built in fields, so that built in fields such as
		// primary key are added to the newly added built in models
		if i == 0 {
			scm.insertBuiltInModels(declarations, oneInputSchemaFile)
		}

		scm.insertBuiltInFields(declarations)

		asts = append(asts, declarations)
	}

	// if we have errors in parsing then no point running validation rules
	if len(parseErrors.Errors) > 0 {
		return nil, parseErrors
	}

	v := validation.NewValidator(asts)
	err := v.RunAllValidators()
	if err != nil {
		return nil, err
	}

	scm.asts = asts

	protoModels := scm.makeProtoModels()
	return protoModels, nil
}

// insertBuiltInFields injects new fields into the parser schema, to represent
// our implicit (or built-in) fields. For example every Model has an <id> field.
func (scm *Builder) insertBuiltInFields(declarations *parser.AST) {
	for _, decl := range declarations.Declarations {
		if decl.Model == nil {
			continue
		}

		fields := []*parser.FieldNode{
			{
				BuiltIn: true,
				Name: parser.NameNode{
					Value: parser.ImplicitFieldNameId,
				},
				Type: parser.TypeNode{
					Value: parser.FieldTypeID,
				},
				Attributes: []*parser.AttributeNode{
					{
						Name: parser.AttributeNameToken{Value: "primaryKey"},
					},
					{
						Name: parser.AttributeNameToken{Value: "default"},
					},
				},
			},
			{
				BuiltIn: true,
				Name: parser.NameNode{
					Value: parser.ImplicitFieldNameCreatedAt,
				},
				Type: parser.TypeNode{
					Value: parser.FieldTypeDatetime,
				},
				Attributes: []*parser.AttributeNode{
					{
						Name: parser.AttributeNameToken{
							Value: "default",
						},
					},
				},
			},
			{
				BuiltIn: true,
				Name: parser.NameNode{
					Value: parser.ImplicitFieldNameUpdatedAt,
				},
				Type: parser.TypeNode{
					Value: parser.FieldTypeDatetime,
				},
				Attributes: []*parser.AttributeNode{
					{
						Name: parser.AttributeNameToken{
							Value: "default",
						},
					},
				},
			},
		}

		var fieldsSection *parser.ModelSectionNode
		for _, section := range decl.Model.Sections {
			if len(section.Fields) > 0 {
				fieldsSection = section
				break
			}
		}

		if fieldsSection == nil {
			decl.Model.Sections = append(decl.Model.Sections, &parser.ModelSectionNode{
				Fields: fields,
			})
		} else {
			fieldsSection.Fields = append(fieldsSection.Fields, fields...)
		}
	}
}

func (scm *Builder) insertBuiltInModels(declarations *parser.AST, schemaFile reader.SchemaFile) {
	declarations.Declarations = append(declarations.Declarations,
		&parser.DeclarationNode{
			Model: &parser.ModelNode{
				BuiltIn: true,
				Name: parser.NameNode{
					Value: "Identity",
					Node: node.Node{
						Pos: lexer.Position{
							Filename: schemaFile.FileName,
						},
					},
				},
			},
		},
	)

	field := &parser.FieldNode{
		BuiltIn: true,
		Name: parser.NameNode{
			Value: "username",
		},
		Type: parser.TypeNode{
			Value: "Text",
		},
	}
	section := &parser.ModelSectionNode{
		Fields: []*parser.FieldNode{field},
	}

	model := declarations.Declarations[len(declarations.Declarations)-1].Model
	model.Sections = append(model.Sections, section)
}
