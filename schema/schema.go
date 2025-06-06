package schema

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/samber/lo"
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

	cfg, err := config.Load(directory)
	if err != nil && config.ToConfigErrors(err) == nil {
		return nil, err
	}

	scm.Config = cfg
	scm.schemaFiles = allInputFiles.SchemaFiles
	return scm.makeFromInputs(allInputFiles)
}

func (scm *Builder) MakeFromString(schemaString string, configString string) (*proto.Schema, error) {
	scm.schemaFiles = append(scm.schemaFiles, &reader.SchemaFile{
		Contents: schemaString,
		FileName: "schema.keel",
	})

	cfg, err := config.LoadFromBytes([]byte(configString), "")
	if err != nil {
		if _, ok := err.(*config.ConfigErrors); !ok {
			// This is a bit messy, but for now we dont return config validation errors from here
			return nil, err
		}
	}

	scm.Config = cfg

	return scm.makeFromInputs(&reader.Inputs{
		SchemaFiles: scm.schemaFiles,
	})
}

// MakeFromFile constructs a proto.Schema from the given inputs.
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
	for i, file := range allInputFiles.SchemaFiles {
		declarations, err := parser.Parse(file)
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
				return nil, parseErrors, fmt.Errorf("parser.Parse() failed on file: %s, with error %v", file.FileName, err)
			}
		}

		if declarations == nil {
			continue
		}

		// Insert built in models like Identity. We only want to call this once
		// so that only one instance of the built in models are added if there
		// are multiple ASTs at play.
		// We want the insertion of built in models to happen
		// before insertion of built in fields, so that built in fields such as
		// primary key are added to the newly added built in models
		if i == 0 {
			scm.insertBuiltInModels(declarations, file)
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
	validationErrors := v.RunAllValidators(false)
	if validationErrors != nil {
		return nil, validationErrors
	}

	scm.asts = asts

	protoModels := scm.makeProtoModels()
	return protoModels, nil
}

// ValidateFromInputs will tyake the given inputs and build the ASTs and run all validators, including/excluding warnings
// based on the given param. Similar with MakeFromInputs, this function avoide building the protoModels for increased
// performance when only validation is required.
func (scm *Builder) ValidateFromInputs(inputs *reader.Inputs, includeWarnings bool) error {
	scm.schemaFiles = inputs.SchemaFiles

	asts, parseErrors, err := scm.PrepareAst(inputs)
	if err != nil {
		return err
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
		return &parseErrors
	}

	v := validation.NewValidator(asts)
	return v.RunAllValidators(includeWarnings)
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
					Value: parser.FieldNameId,
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
					Value: parser.FieldNameCreatedAt,
				},
				Type: parser.NameNode{
					Value: parser.FieldTypeTimestamp,
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
					Value: parser.FieldNameUpdatedAt,
				},
				Type: parser.NameNode{
					Value: parser.FieldTypeTimestamp,
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
		fkFieldsToAdd := map[int]*parser.FieldNode{}

		for i, field := range query.ModelFields(mdl) {
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

			fkFieldsToAdd[i] = fkField
		}

		// Add the new FK fields to the existing model's fields section at the same location as the model fields.
		offset := 1
		keys := lo.Keys(fkFieldsToAdd)
		slices.Sort(keys)
		for _, section := range mdl.Sections {
			if section.Fields != nil {
				for _, v := range keys {
					section.Fields = slices.Insert(section.Fields, v+offset, fkFieldsToAdd[v])
					offset++
				}
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
				Value: parser.IdentityModelName,
				Node: node.Node{
					Pos: lexer.Position{
						Filename: schemaFile.FileName,
					},
				},
			},
		},
	}

	falseLiteral, _ := parser.ParseExpression("false")

	identityFields := []*parser.FieldNode{
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameEmail,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameEmailVerified,
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
							Expression: falseLiteral,
						},
					},
				},
			},
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNamePassword,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypePassword,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameExternalId,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameIssuer,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameName,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameGivenName,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameFamilyName,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameMiddleName,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameNickName,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameProfile,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNamePicture,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameWebsite,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameGender,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameZoneInfo,
			},
			Type: parser.NameNode{
				Value: parser.FieldTypeText,
			},
			Optional: true,
		},
		{
			BuiltIn: true,
			Name: parser.NameNode{
				Value: parser.IdentityFieldNameLocale,
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
		Attribute: scm.emailUniqueAttributeNode(),
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

func (scm *Builder) emailUniqueAttributeNode() *parser.AttributeNode {
	operands := []string{"email", "issuer"}

	if scm.Config != nil {
		for _, c := range scm.Config.Auth.Claims {
			if !c.Unique {
				continue
			}

			operands = append(operands, c.Field)
		}
	}

	operandsExpr, _ := parser.ParseExpression(fmt.Sprintf("[%s]", strings.Join(operands, ",")))

	return &parser.AttributeNode{
		Name: parser.AttributeNameToken{
			Value: parser.AttributeUnique,
		},
		Arguments: []*parser.AttributeArgumentNode{
			{
				Expression: operandsExpr,
			},
		},
	}
}
