package schema

import (
	"fmt"

	"github.com/teamkeel/keel/inputs"
	"github.com/teamkeel/keel/parser"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/validation"
)

// A Schema knows how to produce a (validated) proto.Schema,
// from a given Keel Schema. Construct one, then call the Make method.
type Schema struct {
	schemaDir string
}

// NewSchema provides a Schema that is ready to have its Make method called.
func NewSchema(schemaDir string) *Schema {
	return &Schema{
		schemaDir: schemaDir,
	}
}

// Make constructs a proto.Schema from the .keel files present in the directory
// given at construction time.
func (scm *Schema) Make() (*proto.Schema, error) {
	// These are the main steps:
	//
	// - Locate and read the files in the directory.
	// - For each of the .keel (schema) files present...
	// 		- Parse to AST
	// 		- Add built-in fields
	// - With the parsed (AST) schemas as a set:
	// 		- Validate them (as a set)
	// 		- Convert the set to a single / aggregate proto model

	allInputFiles, err := inputs.Assemble(scm.schemaDir)
	if err != nil {
		return nil, fmt.Errorf("Error assembling input files: %v", err)
	}
	validationInputs := []validation.Input{}
	for _, oneInputSchemaFile := range allInputFiles.SchemaFiles {
		declarations, err := parser.Parse(oneInputSchemaFile.Contents)
		if err != nil {
			return nil, fmt.Errorf("parser.Parse() failed on file: %s, with error %v", oneInputSchemaFile.FileName, err)
		}
		scm.insertBuiltInFields(declarations)
		validationInputs = append(validationInputs, validation.Input{
			FileName: oneInputSchemaFile.FileName,
			ParsedSchema: declarations,
		})
	}
	

	v := validation.NewValidator(validationInputs)
	err = v.RunAllValidators()
	if err != nil {
		return nil, fmt.Errorf("RunAllValidators() failed with: %v", err)
	}

	validatedSchemas := []*parser.Schema{}
	for _, vs := range validationInputs {
		validatedSchemas = append(validatedSchemas, vs.ParsedSchema)
	}
	protoModels := scm.makeProtoModels(validatedSchemas)
	return protoModels, nil
}


// insertBuiltInFields injects new fields into the parser schema, to represent
// our implicit (or built-in) fields. For example every Model has an <id> field.
func (scm *Schema) insertBuiltInFields(declarations *parser.Schema) {
	for _, decl := range declarations.Declarations {
		if decl.Model == nil {
			continue
		}
		field := &parser.ModelField{
			BuiltIn: true,
			Name: parser.ImplicitFieldNameId,
			Type:    parser.FieldTypeID,
			Attributes: []*parser.Attribute{
				{
					Name: "primaryKey",
				},
			},
		}
		section := &parser.ModelSection{
			Fields: []*parser.ModelField{field},
		}
		model := decl.Model
		model.Sections = append(model.Sections, section)
	}
}

