package schema

import (
	"fmt"
	"io/ioutil"

	"github.com/teamkeel/keel/inputs"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation"
)

// A Schema knows how to produce a (validated) proto.Schema,
// from a given Keel Schema. Construct one, then call the Make method.
type Schema struct {
}

// MakeFromDirectory constructs a proto.Schema from the .keel files present in the given
// directory.
func (scm *Schema) MakeFromDirectory(directory string) (*proto.Schema, error) {
	allInputFiles, err := inputs.Assemble(directory)
	if err != nil {
		return nil, fmt.Errorf("error assembling input files: %v", err)
	}
	schema, err := scm.makeFromInputs(allInputFiles)
	if err != nil {
		verrs, ok := err.(validation.ValidationErrors)
		if ok {
			return nil, verrs
		} else {
			return nil, fmt.Errorf("error reading file: %v", err)
		}
	}

	return schema, nil
}

// MakeFromFile constructs a proto.Schema from the given .keel file.
func (scm *Schema) MakeFromFile(filename string) (*proto.Schema, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	schemaFile := inputs.SchemaFile{
		FileName: filename,
		Contents: string(fileBytes),
	}
	allInputFiles := &inputs.Inputs{
		Directory:   "Unspecified",
		SchemaFiles: []inputs.SchemaFile{schemaFile},
	}
	schema, err := scm.makeFromInputs(allInputFiles)
	if err != nil {
		verrs, ok := err.(validation.ValidationErrors)
		if ok {
			return nil, verrs
		} else {
			return nil, fmt.Errorf("error reading file: %v", err)
		}
	}

	return schema, nil
}

func (scm *Schema) makeFromInputs(allInputFiles *inputs.Inputs) (*proto.Schema, error) {
	// - For each of the .keel (schema) files specified...
	// 		- Parse to AST
	// 		- Add built-in fields
	// - With the parsed (AST) schemas as a set:
	// 		- Validate them (as a set)
	// 		- Convert the set to a single / aggregate proto model
	validationInputs := []validation.Input{}
	for _, oneInputSchemaFile := range allInputFiles.SchemaFiles {
		declarations, err := parser.Parse(&oneInputSchemaFile)
		if err != nil {
			return nil, fmt.Errorf("parser.Parse() failed on file: %s, with error %v", oneInputSchemaFile.FileName, err)
		}
		scm.insertBuiltInFields(declarations)
		validationInputs = append(validationInputs, validation.Input{
			FileName:     oneInputSchemaFile.FileName,
			ParsedSchema: declarations,
		})
	}

	v := validation.NewValidator(validationInputs)
	err := v.RunAllValidators()
	if err != nil {
		return nil, err
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
			Name:    parser.ImplicitFieldNameId,
			Type:    parser.FieldTypeID,
			Attributes: []*parser.Attribute{
				{
					Name: "primaryKey",
				},
				{
					Name: "unique",
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
