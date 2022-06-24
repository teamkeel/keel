package schema

import (
	"fmt"

	"github.com/teamkeel/keel/proto"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/schema/validation"
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
	for _, oneInputSchemaFile := range allInputFiles.SchemaFiles {
		declarations, err := parser.Parse(&oneInputSchemaFile)
		if err != nil {
			return nil, fmt.Errorf("parser.Parse() failed on file: %s, with error %v", oneInputSchemaFile.FileName, err)
		}
		scm.insertBuiltInFields(declarations)
		asts = append(asts, declarations)
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
				Type: parser.FieldTypeID,
				Attributes: []*parser.AttributeNode{
					{
						Name: parser.AttributeNameToken{Value: "primaryKey"},
					},
				},
			},
			{
				BuiltIn: true,
				Name: parser.NameNode{
					Value: parser.ImplicitFieldNameCreatedAt,
				},
				Type: parser.FieldTypeDatetime,
				// TODO: add @default(now())
			},
			{
				BuiltIn: true,
				Name: parser.NameNode{
					Value: parser.ImplicitFieldNameUpdatedAt,
				},
				Type: parser.FieldTypeDatetime,
				// TODO: add default(now())
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
