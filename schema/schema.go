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
	schemaFiles []*reader.SchemaFile
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

func (scm *Builder) MakeFromString(schemaString string, configString string) (*proto.Schema, error) {
	scm.schemaFiles = append(scm.schemaFiles, &reader.SchemaFile{
		Contents: schemaString,
		FileName: "schema.keel",
	})

	config, err := config.LoadFromBytes([]byte(configString))
	if err != nil {
		return nil, err
	}

	scm.Config = config

	return scm.makeFromInputs(&reader.Inputs{
		SchemaFiles: scm.schemaFiles,
	})
}

// MakeFromFile constructs a proto.Schema from the given inputs
func (scm *Builder) MakeFromInputs(inputs *reader.Inputs) (*proto.Schema, error) {
	scm.schemaFiles = inputs.SchemaFiles

	return scm.makeFromInputs(inputs)
}

func (scm *Builder) SchemaFiles() []*reader.SchemaFile {
	return scm.schemaFiles
}

func (scm *Builder) ASTs() []*parser.AST {
	return scm.asts
}

// PrepareAst will parse the ASTs and will add built-in models, fields, and other bits.
func (scm *Builder) PrepareAst(allInputFiles *reader.Inputs) ([]*parser.AST, errorhandling.ValidationErrors, error) {
	asts := []*parser.AST{}
	parseErrors := errorhandling.ValidationErrors{}

	// - For each of the .keel (schema) files specified...
	// 		- Parse to AST
	// 		- Add built-in fields
	for i, oneInputSchemaFile := range allInputFiles.SchemaFiles {
		declarations, err := parser.Parse(oneInputSchemaFile)
		if err != nil {
			// try to convert into a validation error and move to next schema file
			if perr, ok := err.(parser.Error); ok {
				verr := errorhandling.NewValidationError(errorhandling.ErrorInvalidSyntax, errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Message": perr.Error(),
					},
				}, perr)
				parseErrors.Errors = append(parseErrors.Errors, verr)
			} else {
				return nil, parseErrors, fmt.Errorf("parser.Parse() failed on file: %s, with error %v", oneInputSchemaFile.FileName, err)
			}
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

	return asts, parseErrors, nil
}

func (scm *Builder) makeFromInputs(allInputFiles *reader.Inputs) (*proto.Schema, error) {
	asts, parseErrors, err := scm.PrepareAst(allInputFiles)
	if err != nil {
		return nil, err
	}

	// insert the foreign key fields (for relationships)
	errDetails := scm.insertForeignKeyFields(asts)
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
	validationErrors := v.RunAllValidators()
	if validationErrors != nil {
		return nil, validationErrors
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
func (scm *Builder) insertForeignKeyFields(asts []*parser.AST) *errorhandling.ErrorDetails {

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

func (scm *Builder) insertBuiltInModels(declarations *parser.AST, schemaFile *reader.SchemaFile) {
	scm.insertIdentityModel(declarations, schemaFile)
}

func (scm *Builder) insertIdentityModel(declarations *parser.AST, schemaFile *reader.SchemaFile) {
	identityModelDeclaration := &parser.DeclarationNode{
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

	identityFields := []*parser.FieldNode{
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameEmail,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameEmailVerified,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeBoolean,
			},
			Optional: false,
			Attributes: []*parser.AttributeNode{
				{
					Name: parser.AttributeNameToken{
						Value: parser.AttributeDefault,
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
														False: true,
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
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNamePassword,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypePassword,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameExternalId,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameIssuer,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameName,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameGivenName,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameFamilyName,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameMiddleName,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameNickName,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameProfile,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNamePicture,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameWebsite,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameGender,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameZoneInfo,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.ImplicitIdentityFieldNameLocale,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
	}

	if scm.Config != nil {
		for _, c := range scm.Config.Auth.Claims {
			identityFields = append(identityFields, &parser.FieldNode{
				BuiltIn: true,
				Name: parser.NameNode{
					Value: c.Field,
				},
				Type: parser.NameNode{
					Value: parser.FieldTypeText,
				},
				Optional: true,
			})
		}
	}

	requestPasswordReset := &parser.ActionNode{
		BuiltIn: true,
		Type:    parser.NameNode{Value: parser.ActionTypeWrite},
		Name:    parser.NameNode{Value: parser.RequestPasswordResetActionName},
		Inputs: []*parser.ActionInputNode{
			{
				Type: parser.Ident{Fragments: []*parser.IdentFragment{{Fragment: "RequestPasswordResetInput"}}}, Optional: false,
			},
		},
		Returns: []*parser.ActionInputNode{
			{
				Type: parser.Ident{Fragments: []*parser.IdentFragment{{Fragment: "RequestPasswordResetResponse"}}}, Optional: false,
			},
		},
	}

	resetPasswordAction := &parser.ActionNode{
		BuiltIn: true,
		Type:    parser.NameNode{Value: parser.ActionTypeWrite},
		Name:    parser.NameNode{Value: parser.PasswordResetActionName},
		Inputs: []*parser.ActionInputNode{
			{
				Type: parser.Ident{Fragments: []*parser.IdentFragment{{Fragment: "ResetPasswordInput"}}}, Optional: false,
			},
		},
		Returns: []*parser.ActionInputNode{
			{
				Type: parser.Ident{Fragments: []*parser.IdentFragment{{Fragment: "ResetPasswordResponse"}}}, Optional: false,
			},
		},
	}

	fieldsSection := &parser.ModelSectionNode{
		Fields: identityFields,
	}

	actionsSection := &parser.ModelSectionNode{
		Actions: []*parser.ActionNode{requestPasswordReset, resetPasswordAction},
	}

	identityModelDeclaration.Model.Sections = append(identityModelDeclaration.Model.Sections, fieldsSection, actionsSection)

	uniqueModelSection := &parser.ModelSectionNode{
		Attribute: emailUniqueAttributeNode(),
	}

	identityModelDeclaration.Model.Sections = append(identityModelDeclaration.Model.Sections, uniqueModelSection)

	resetPasswordInputDeclaration := &parser.DeclarationNode{
		Message: &parser.MessageNode{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: "ResetPasswordInput",
			},
			Fields: []*parser.FieldNode{
				{
					Name: parser.NameNode{
						Value: "token",
					},
					Type: parser.NameNode{
						Value: "Text",
					},
				},
				{
					Name: parser.NameNode{
						Value: "password",
					},
					Type: parser.NameNode{
						Value: "Text",
					},
				},
			},
		},
	}

	resetPasswordResponseDeclaration := &parser.DeclarationNode{
		Message: &parser.MessageNode{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: "ResetPasswordResponse",
			},
			Fields: []*parser.FieldNode{},
		},
	}

	requestPasswordResetInputDeclaration := &parser.DeclarationNode{
		Message: &parser.MessageNode{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: "RequestPasswordResetInput",
			},
			Fields: []*parser.FieldNode{
				{
					Name: parser.NameNode{
						Value: "email",
					},
					Type: parser.NameNode{
						Value: "Text",
					},
				},
				{
					Name: parser.NameNode{
						Value: "redirectUrl",
					},
					Type: parser.NameNode{
						Value: "Text",
					},
				},
			},
		},
	}

	requestPasswordResetResponseDeclaration := &parser.DeclarationNode{
		Message: &parser.MessageNode{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: "RequestPasswordResetResponse",
			},
			Fields: []*parser.FieldNode{},
		},
	}

	declarations.Declarations = append(
		declarations.Declarations,
		identityModelDeclaration,
		requestPasswordResetInputDeclaration,
		requestPasswordResetResponseDeclaration,
		resetPasswordInputDeclaration,
		resetPasswordResponseDeclaration)
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
