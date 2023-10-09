package schema

import (
	"errors"
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"

	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/schema/validation"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// A Builder knows how to produce a (validated) proto.Schema,
// from a given Keel Builder. Construct one, then call the Make method.
type Builder struct {
	asts        []*parser.AST
	schemaFiles []reader.SchemaFile
	Config      *config.ProjectConfig
	proto       *proto.Schema
}

var ErrNoSchemaFiles = errors.New("no schema files found")

// MakeFromDirectory constructs a proto.Schema from the .keel files present in the given
// directory.
func (scm *Builder) MakeFromDirectory(directory string) (*proto.Schema, error) {
	allInputFiles, err := reader.FromDir(directory)
	if err != nil {
		return nil, err
	}

	if len(allInputFiles.SchemaFiles) == 0 {
		return nil, ErrNoSchemaFiles
	}

	config, err := config.Load(directory)
	if err != nil {
		return nil, err
	}

	scm.Config = config
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

func (scm *Builder) MakeFromString(schemaString string) (*proto.Schema, error) {
	scm.schemaFiles = append(scm.schemaFiles, reader.SchemaFile{
		Contents: schemaString,
		FileName: "schema.keel",
	})

	return scm.makeFromInputs(&reader.Inputs{
		SchemaFiles: scm.schemaFiles,
	})
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

		// This inserts the built in fields like "createdAt" etc. But it does not insert
		// the relationship foreign key fields, because we need to defer that until all the
		// models in the global set have been captured and modelled.
		scm.insertBuiltInFields(declarations)

		// Add environment variables to the ASTs
		scm.addEnvironmentVariables(declarations)

		// Add secrets to the ASTs
		scm.addSecrets(declarations)

		asts = append(asts, declarations)
	}

	// All the code below this point - depends on having access to the global
	// i.e. aggregated asts from multiple files. Mostly in order to be able to
	// reason over ALL models scope.

	// Inject implied reverse relationship fields into the Identity model.
	// This creates our "backlinks" feature from the Identity model.
	errDetails := scm.insertAllBackLinkFields(asts)
	if errDetails != nil {
		parseErrors.Errors = append(parseErrors.Errors, &errorhandling.ValidationError{
			ErrorDetails: errDetails,
		})
	}

	// Now insert the foreign key fields (for relationships)
	errDetails = scm.insertForeignKeyFields(asts)
	if errDetails != nil {
		parseErrors.Errors = append(parseErrors.Errors, &errorhandling.ValidationError{
			ErrorDetails: errDetails,
		})
	}

	// if we have errors in parsing then no point running validation rules
	if len(parseErrors.Errors) > 0 {
		return nil, &parseErrors
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
				Type: parser.NameNode{
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
				Type: parser.NameNode{
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
				Type: parser.NameNode{
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

// insertForeignKeyFields works with the given GLOBAL set of asts, i.e. a set that has been
// built and combined from all input files. It analyses the foreign key fields that should be auto
// generated and injected into each model.
func (scm *Builder) insertForeignKeyFields(
	asts []*parser.AST) *errorhandling.ErrorDetails {

	for _, mdl := range query.Models(asts) {
		fkFieldsToAdd := []*parser.FieldNode{}

		for _, field := range query.ModelFields(mdl) {
			if query.Model(asts, field.Type.Value) == nil {
				continue
			}

			candidates := query.GetRelationshipCandidates(asts, mdl, field)

			if candidates == nil || len(candidates) != 1 {
				continue
			}

			relationship := candidates[0]

			if !(relationship.Field == nil ||
				query.ValidOneToHasMany(field, relationship.Field) ||
				query.ValidUniqueOneToHasOne(field, relationship.Field)) {
				continue
			}
			// if !(query.IsHasOneModelField(asts, field) || query.FieldIsUnique(field)) {
			// 	continue
			// }

			// if !query.IsHasOneModelField(asts, field) || query.FieldIsUnique(field)) {
			// 	continue
			// }

			// otherModel := query.Model(asts, field.Type.Value)
			// if otherModel == nil {
			// 	continue
			// }

			//if !query.ValidOneToHasMany(field, )

			referredToModelName := casing.ToCamel(field.Type.Value)
			referredToModel := query.Model(asts, referredToModelName)

			if referredToModel == nil {
				errDetails := &errorhandling.ErrorDetails{
					Message: fmt.Sprintf("cannot find the model referred to (%s) by field %s, on model %s",
						referredToModelName, field.Name, mdl.Name),
					Hint: "make sure you declare this model",
				}
				return errDetails
			}

			referredToModelPK := query.PrimaryKey(referredToModelName, asts)

			// This is the single source of truth for how we name foreign key fields.
			// Later on, we'll let the the user name them in the schema language.
			generatedForeignKeyName := field.Name.Value + casing.ToCamel(referredToModelPK.Name.Value)

			fkField := &parser.FieldNode{
				BuiltIn:  true,
				Optional: field.Optional,
				Name: parser.NameNode{
					Value: generatedForeignKeyName,
				},
				Type: parser.NameNode{
					Value: parser.FieldTypeID,
				},
				Attributes: []*parser.AttributeNode{},
			}

			// Give the FK field the same "uniqueness" as the "owning" field.
			if query.FieldHasAttribute(field, parser.AttributeUnique) {
				attr := parser.AttributeNode{
					Name: parser.AttributeNameToken{Value: parser.AttributeUnique},
				}
				fkField.Attributes = append(fkField.Attributes, &attr)
			}

			fkFieldsToAdd = append(fkFieldsToAdd, fkField)
		}
		// Add the new FK fields to the existing model's fields section.
		for _, section := range mdl.Sections {
			if section.Fields != nil {
				section.Fields = append(section.Fields, fkFieldsToAdd...)
			}
		}
	}
	return nil
}

func (scm *Builder) insertBuiltInModels(declarations *parser.AST, schemaFile reader.SchemaFile) {
	scm.insertIdentityModel(declarations, schemaFile)
}

func (scm *Builder) insertIdentityModel(declarations *parser.AST, schemaFile reader.SchemaFile) {
	declaration := &parser.DeclarationNode{
		Model: &parser.ModelNode{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityModelName,
				Node: node.Node{
					Pos: lexer.Position{
						Filename: schemaFile.FileName,
					},
				},
			},
		},
	}

	emailField := &parser.FieldNode{
		BuiltIn: true,
		Name: parser.NameNode{
			Value: parser.ImplicitIdentityFieldNameEmail,
		},
		Type: parser.NameNode{
			Value: parser.FieldTypeText,
		},
		Optional: true,
	}

	defaultAttributeArgs := []*parser.AttributeArgumentNode{}

	defaultAttributeArgs = append(defaultAttributeArgs, &parser.AttributeArgumentNode{
		Expression: &parser.Expression{
			Or: []*parser.OrExpression{
				{
					And: []*parser.ConditionWrap{
						{
							Condition: &parser.Condition{
								LHS: &parser.Operand{
									False: true,
								},
							},
						},
					},
				},
			},
		},
	})

	defaultAttribute := &parser.AttributeNode{
		Name: parser.AttributeNameToken{
			Value: parser.AttributeDefault,
		},
		Arguments: defaultAttributeArgs,
	}

	emailVerifiedField := &parser.FieldNode{
		BuiltIn: true,
		Name: parser.NameNode{
			Value: parser.ImplicitIdentityFieldNameEmailVerified,
		},
		Type: parser.NameNode{
			Value: parser.FieldTypeBoolean,
		},
		Optional:   false,
		Attributes: []*parser.AttributeNode{defaultAttribute},
	}

	passwordField := &parser.FieldNode{
		BuiltIn: true,
		Name: parser.NameNode{
			Value: parser.ImplicitIdentityFieldNamePassword,
		},
		Type: parser.NameNode{
			Value: parser.FieldTypePassword,
		},
		Optional: true,
	}

	externalIdField := &parser.FieldNode{
		BuiltIn: true,
		Name: parser.NameNode{
			Value: parser.ImplicitIdentityFieldNameExternalId,
		},
		Type: parser.NameNode{
			Value: parser.FieldTypeText,
		},
		Optional: true,
	}
	issuerField := &parser.FieldNode{
		BuiltIn: true,
		Name: parser.NameNode{
			Value: parser.ImplicitIdentityFieldNameIssuer,
		},
		Type: parser.NameNode{
			Value: parser.FieldTypeText,
		},

		Optional: true,
	}

	section := &parser.ModelSectionNode{
		Fields: []*parser.FieldNode{emailField, emailVerifiedField, passwordField, externalIdField, issuerField},
	}

	declaration.Model.Sections = append(declaration.Model.Sections, section)

	uniqueModelSection := &parser.ModelSectionNode{
		Attribute: emailUniqueAttributeNode(),
	}

	declaration.Model.Sections = append(declaration.Model.Sections, uniqueModelSection)

	declarations.Declarations = append(declarations.Declarations, declaration)

	// Making the identity model and operations available on all APIs

	// Note this only applies to API's that the user has defined in their schema.
	// You could say we "sneak the Identity model in to those APIs"/
	//
	// However, if the schema doesn't have any API's defined - there is code elsewhere that creates
	// a default API for you - and that includes ALL models (which include the new auto generated Audit model.)
	for _, d := range declarations.Declarations {
		if d.API != nil {
			for _, s := range d.API.Sections {
				if s.Models != nil {
					s.Models = append(s.Models, &parser.ModelsNode{
						Name: parser.NameNode{
							Value: parser.ImplicitIdentityModelName,
						},
					})
				}
			}
		}
	}
}

func (scm *Builder) addEnvironmentVariables(declarations *parser.AST) {
	if scm.Config == nil {
		return
	}

	declarations.EnvironmentVariables = append(declarations.EnvironmentVariables, scm.Config.AllEnvironmentVariables()...)
}

func (scm *Builder) addSecrets(declarations *parser.AST) {
	if scm.Config == nil {
		return
	}

	declarations.Secrets = append(declarations.Secrets, scm.Config.AllSecrets()...)
}

func emailUniqueAttributeNode() *parser.AttributeNode {
	// we need to add a composite unique attribute on email + issuer
	// unfortunately this is the only way to do it prior to proto generation
	return &parser.AttributeNode{
		Name: parser.AttributeNameToken{
			Value: parser.AttributeUnique,
		},
		Arguments: []*parser.AttributeArgumentNode{
			{
				Expression: &parser.Expression{
					Or: []*parser.OrExpression{
						{
							And: []*parser.ConditionWrap{
								{
									Condition: &parser.Condition{
										LHS: &parser.Operand{
											Array: &parser.Array{
												Values: []*parser.Operand{
													{
														Ident: &parser.Ident{
															Fragments: []*parser.IdentFragment{
																{
																	Fragment: "email",
																},
															},
														},
													},
													{
														Ident: &parser.Ident{
															Fragments: []*parser.IdentFragment{
																{
																	Fragment: "issuer",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
