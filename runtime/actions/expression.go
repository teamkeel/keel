package actions

import (
	"context"
	"fmt"
	"net/textproto"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
)

// Constructs and adds an LEFT JOIN from a splice of fragments (representing an operand in an expression or implicit input).
// The fragment slice must include the base model as the first item, for example: "post." in post.author.publisher.isActive
func (query *QueryBuilder) AddJoinFromFragments(schema *proto.Schema, fragments []string) error {
	if fragments[0] == "ctx" {
		return nil
	}

	fragments, err := normalisedFragments(schema, fragments)
	if err != nil {
		return err
	}

	model := casing.ToCamel(fragments[0])
	fragmentCount := len(fragments)

	for i := 1; i < fragmentCount-1; i++ {
		currentFragment := fragments[i]

		if !proto.ModelHasField(schema, model, currentFragment) {
			return fmt.Errorf("this model: %s, does not have a field of name: %s", model, currentFragment)
		}

		// We know that the current fragment is a related model because it's not the last fragment
		relatedModelField := proto.FindField(schema.Models, model, currentFragment)
		relatedModel := relatedModelField.Type.ModelName.Value
		foreignKeyField := proto.GetForeignKeyFieldName(schema.Models, relatedModelField)
		primaryKey := "id"

		var leftOperand *QueryOperand
		var rightOperand *QueryOperand

		switch {
		case relatedModelField.IsBelongsTo():
			// In a "belongs to" the foreign key is on _this_ model
			leftOperand = ExpressionField(fragments[:i+1], primaryKey, false)
			rightOperand = ExpressionField(fragments[:i], foreignKeyField, false)
		default:
			// In all others the foreign key is on the _other_ model
			leftOperand = ExpressionField(fragments[:i+1], foreignKeyField, false)
			rightOperand = ExpressionField(fragments[:i], primaryKey, false)
		}

		query.Join(relatedModel, leftOperand, rightOperand)

		model = relatedModelField.Type.ModelName.Value
	}

	return nil
}

func generateOperand(ctx context.Context, schema *proto.Schema, model *proto.Model, action *proto.Action, inputs map[string]any, fragments []string) (*QueryOperand, error) {
	ident, err := normalisedFragments(schema, fragments)
	if err != nil {
		return nil, err
	}

	var queryOperand *QueryOperand
	switch {
	case len(ident) == 2 && proto.EnumExists(schema.Enums, ident[0]):
		return Value(ident[1]), nil
	case expressions.IsModelDbColumn(model, ident):
		var err error
		queryOperand, err = operandFromFragments(schema, ident)
		if err != nil {
			return nil, err
		}
	case expressions.IsInput(schema, action, ident):
		value, ok := inputs[ident[0]]
		if !ok {
			return nil, fmt.Errorf("implicit or explicit input '%s' does not exist in arguments", ident[0])
		}
		return Value(value), nil
	case expressions.IsContextDbColumn(ident):
		// If this is a value from ctx that requires a database read (such as with identity backlinks),
		// then construct an inline query for this operand.  This is necessary because we can't retrieve this value
		// from the current query builder.

		// Remove the ctx fragment
		ident = ident[1:]

		identityModel := schema.FindModel(strcase.ToCamel(ident[0]))

		identityId := ""
		if auth.IsAuthenticated(ctx) {
			identity, err := auth.GetIdentity(ctx)
			if err != nil {
				return nil, err
			}
			identityId = identity[parser.FieldNameId].(string)
		}

		query := NewQuery(identityModel)
		err := query.AddJoinFromFragments(schema, ident)
		if err != nil {
			return nil, err
		}

		err = query.Where(IdField(), Equals, Value(identityId))
		if err != nil {
			return nil, err
		}

		selectField, err := operandFromFragments(schema, ident)
		if err != nil {
			return nil, err
		}

		// If there are no matches in the subquery then null will be returned, but null
		// will cause IN and NOT IN filtering of this subquery result to always evaluate as false.
		// Therefore we need to filter out null.
		query.And()
		err = query.Where(selectField, NotEquals, Null())
		if err != nil {
			return nil, err
		}

		if selectField.IsArrayField() {
			query.SelectUnnested(selectField)
		} else {
			query.Select(selectField)
		}

		queryOperand = InlineQuery(query, selectField)

	case expressions.IsContextIdentityId(ident):
		isAuthenticated := auth.IsAuthenticated(ctx)
		if !isAuthenticated {
			queryOperand = Null()
		} else {
			identity, err := auth.GetIdentity(ctx)
			if err != nil {
				return nil, err
			}
			queryOperand = Value(identity[parser.FieldNameId].(string))
		}
	case expressions.IsContextIsAuthenticatedField(ident):
		isAuthenticated := auth.IsAuthenticated(ctx)
		queryOperand = Value(isAuthenticated)
	case expressions.IsContextNowField(ident):
		queryOperand = Value(runtimectx.GetNow())
	case expressions.IsContextEnvField(ident):
		envVarName := ident[2]
		queryOperand = Value(os.Getenv(envVarName))
	case expressions.IsContextSecretField(ident):
		secret, err := runtimectx.GetSecret(ctx, ident[2])
		if err != nil {
			return nil, err
		}
		queryOperand = Value(secret)
	case expressions.IsContextHeadersField(ident):
		headerName := ident[2]

		// First we parse the header name to kebab. MyCustomHeader will become my-custom-header.
		kebab := strcase.ToKebab(headerName)

		// Then get canonical name. my-custom-header will become My-Custom-Header.
		// https://pkg.go.dev/net/http#Header.Get
		canonicalName := textproto.CanonicalMIMEHeaderKey(kebab)

		headers, err := runtimectx.GetRequestHeaders(ctx)
		if err != nil {
			return nil, err
		}
		if value, ok := headers[canonicalName]; ok {
			queryOperand = Value(strings.Join(value, ", "))
		} else {
			queryOperand = Value("")
		}
	default:
		return nil, fmt.Errorf("cannot handle fragments: %s", strings.Join(ident, "."))
	}

	return queryOperand, nil
}

func normalisedFragments(schema *proto.Schema, fragments []string) ([]string, error) {
	isModelField := false
	isCtx := fragments[0] == "ctx"

	if isCtx {
		// We dont bother normalising ctx.isAuthenticated, ctx.secrets, etc.
		if fragments[1] != "identity" {
			return fragments, nil
		}

		// If this is a context backlink, then remove the first "ctx" fragment.
		fragments = fragments[1:]
	}

	// The first fragment will always be the root model name, e.g. "author" in author.posts.title
	modelTarget := schema.FindModel(casing.ToCamel(fragments[0]))
	if modelTarget == nil {
		// If it's not the model, then it could be an input
		return fragments, nil
		//return nil, fmt.Errorf("model '%s' does not exist in schema", casing.ToCamel(fragments[0]))
	}

	var fieldTarget *proto.Field
	for i := 1; i < len(fragments); i++ {
		fieldTarget = proto.FindField(schema.Models, modelTarget.Name, fragments[i])
		if fieldTarget.Type.Type == proto.Type_TYPE_MODEL {
			modelTarget = schema.FindModel(fieldTarget.Type.ModelName.Value)
			if modelTarget == nil {
				return nil, fmt.Errorf("model '%s' does not exist in schema", fieldTarget.Type.ModelName.Value)
			}
		}
	}

	// If no field is provided, for example: @where(account in ...)
	// Or if the target field is a MODEL, for example:
	if fieldTarget == nil || fieldTarget.Type.Type == proto.Type_TYPE_MODEL {
		isModelField = true
	}

	if isModelField && len(fragments) == 1 {
		// One fragment is only possible if the expression is only referencing the model.
		// For example, @where(account in ...)
		// Add a new fragment 'id'
		fragments = append(fragments, parser.FieldNameId)
	} else if isModelField {
		i := 0
		if fragments[0] == "ctx" {
			i++
		}

		modelTarget := schema.FindModel(casing.ToCamel(fragments[i]))
		if modelTarget == nil {
			return nil, fmt.Errorf("model '%s' does not exist in schema", casing.ToCamel(fragments[i]))
		}

		var fieldTarget *proto.Field
		for i := i + 1; i < len(fragments); i++ {
			fieldTarget = proto.FindField(schema.Models, modelTarget.Name, fragments[i])
			if fieldTarget.Type.Type == proto.Type_TYPE_MODEL {
				modelTarget = schema.FindModel(fieldTarget.Type.ModelName.Value)
				if modelTarget == nil {
					return nil, fmt.Errorf("model '%s' does not exist in schema", fieldTarget.Type.ModelName.Value)
				}
			}
		}

		if fieldTarget.IsHasOne() || fieldTarget.IsHasMany() {
			// Add a new fragment 'id'
			fragments = append(fragments, parser.FieldNameId)
		} else {
			// Replace the last fragment with the foreign key field
			fragments[len(fragments)-1] = fmt.Sprintf("%sId", fragments[len(fragments)-1])
		}
	}

	if isCtx {
		fragments = append([]string{"ctx"}, fragments...)
	}

	return fragments, nil
}

// Constructs a QueryOperand from a splice of fragments, representing an expression operand or implicit input.
// The fragment slice must include the base model as the first fragment, for example: post.author.publisher.isActive
func operandFromFragments(schema *proto.Schema, fragments []string) (*QueryOperand, error) {
	fragments, err := normalisedFragments(schema, fragments)
	if err != nil {
		return nil, err
	}

	var field string
	model := casing.ToCamel(fragments[0])
	fragmentCount := len(fragments)
	isArray := false

	for i := 1; i < fragmentCount; i++ {
		currentFragment := fragments[i]

		if !proto.ModelHasField(schema, model, currentFragment) {
			return nil, fmt.Errorf("this model: %s, does not have a field of name: %s", model, currentFragment)
		}

		if i < fragmentCount-1 {
			// We know that the current fragment is a model because it's not the last fragment
			relatedModelField := proto.FindField(schema.Models, model, currentFragment)
			model = relatedModelField.Type.ModelName.Value
		} else {
			// The last fragment is referencing the field
			field = currentFragment
			isArray = proto.FindField(schema.Models, model, currentFragment).Type.Repeated

		}
	}

	return ExpressionField(fragments[:len(fragments)-1], field, isArray), nil
}
