package actions

import (
	"context"
	"errors"
	"time"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func FindIdentityById(ctx context.Context, schema *proto.Schema, id string) (*auth.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)
	query := NewQuery(identityModel)
	err := query.Where(IdField(), Equals, Value(id))
	if err != nil {
		return nil, err
	}

	query.AppendSelect(AllFields())
	result, err := query.SelectStatement().ExecuteToSingle(ctx)

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return mapToIdentity(result)
}

func FindIdentityByEmail(ctx context.Context, schema *proto.Schema, email string, issuer string) (*auth.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)
	query := NewQuery(identityModel)
	err := query.Where(Field("email"), Equals, Value(email))
	if err != nil {
		return nil, err
	}
	query.And()
	err = query.Where(Field("issuer"), Equals, Value(issuer))
	if err != nil {
		return nil, err
	}

	query.AppendSelect(AllFields())
	result, err := query.SelectStatement().ExecuteToSingle(ctx)

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return mapToIdentity(result)
}

func FindIdentityByExternalId(ctx context.Context, schema *proto.Schema, externalId string, issuer string) (*auth.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)
	query := NewQuery(identityModel)
	err := query.Where(Field("externalId"), Equals, Value(externalId))
	if err != nil {
		return nil, err
	}
	query.And()
	err = query.Where(Field("issuer"), Equals, Value(issuer))
	if err != nil {
		return nil, err
	}

	query.AppendSelect(AllFields())
	result, err := query.SelectStatement().ExecuteToSingle(ctx)

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return mapToIdentity(result)
}

func CreateIdentity(ctx context.Context, schema *proto.Schema, email string, password string, issuer string) (*auth.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(identityModel)
	query.AddWriteValues(map[string]*QueryOperand{
		"email":    Value(email),
		"password": Value(password),
		"issuer":   Value(issuer),
	})
	query.AppendSelect(AllFields())
	query.AppendReturning(IdField())

	result, err := query.InsertStatement(ctx).ExecuteToSingle(ctx)
	if err != nil {
		return nil, err
	}

	return mapToIdentity(result)
}

func CreateIdentityWithClaims(ctx context.Context, schema *proto.Schema, externalId string, issuer string, standardClaims *oauth.IdTokenClaims, customClaims map[string]any) (*auth.Identity, error) {
	ctx, span := tracer.Start(ctx, "Create Identity")
	defer span.End()

	span.SetAttributes(attribute.String("externalId", externalId))
	span.SetAttributes(attribute.String("issuer", issuer))

	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(identityModel)

	query.AddWriteValues(map[string]*QueryOperand{
		// default 'email' scope claims
		parser.ImplicitIdentityFieldNameExternalId: Value(externalId),
		parser.ImplicitIdentityFieldNameIssuer:     Value(issuer),

		// default 'profile' scope claims
		parser.ImplicitIdentityFieldNameEmail:         Value(standardClaims.Email),
		parser.ImplicitIdentityFieldNameEmailVerified: Value(standardClaims.EmailVerified),

		// default 'profile' scope claims
		parser.ImplicitIdentityFieldNameName:       ValueOrNullIfEmpty(standardClaims.Name),
		parser.ImplicitIdentityFieldNameGivenName:  ValueOrNullIfEmpty(standardClaims.GivenName),
		parser.ImplicitIdentityFieldNameFamilyName: ValueOrNullIfEmpty(standardClaims.FamilyName),
		parser.ImplicitIdentityFieldNameMiddleName: ValueOrNullIfEmpty(standardClaims.MiddleName),
		parser.ImplicitIdentityFieldNameNickName:   ValueOrNullIfEmpty(standardClaims.NickName),
		parser.ImplicitIdentityFieldNameProfile:    ValueOrNullIfEmpty(standardClaims.Profile),
		parser.ImplicitIdentityFieldNamePicture:    ValueOrNullIfEmpty(standardClaims.Picture),
		parser.ImplicitIdentityFieldNameWebsite:    ValueOrNullIfEmpty(standardClaims.Website),
		parser.ImplicitIdentityFieldNameGender:     ValueOrNullIfEmpty(standardClaims.Gender),
		parser.ImplicitIdentityFieldNameZoneInfo:   ValueOrNullIfEmpty(standardClaims.ZoneInfo),
		parser.ImplicitIdentityFieldNameLocale:     ValueOrNullIfEmpty(standardClaims.Locale),
	})

	for k, v := range customClaims {
		query.AddWriteValue(Field(k), ValueOrNullIfEmpty(v))
	}

	query.AppendSelect(AllFields())
	query.AppendReturning(IdField())

	result, err := query.InsertStatement(ctx).ExecuteToSingle(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapToIdentity(result)
}

func UpdateIdentityWithClaims(ctx context.Context, schema *proto.Schema, externalId string, issuer string, standardClaims *oauth.IdTokenClaims, customClaims map[string]any) (*auth.Identity, error) {
	ctx, span := tracer.Start(ctx, "Update Identity")
	defer span.End()

	span.SetAttributes(attribute.String("externalId", standardClaims.Subject))
	span.SetAttributes(attribute.String("issuer", standardClaims.Issuer))

	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(identityModel)

	err := query.Where(Field("externalId"), Equals, Value(standardClaims.Subject))
	if err != nil {
		return nil, err
	}
	query.And()
	err = query.Where(Field("issuer"), Equals, Value(standardClaims.Issuer))
	if err != nil {
		return nil, err
	}

	query.AddWriteValues(map[string]*QueryOperand{
		// default 'email' scope claims
		parser.ImplicitIdentityFieldNameEmail:         Value(standardClaims.Email),
		parser.ImplicitIdentityFieldNameEmailVerified: Value(standardClaims.EmailVerified),

		// default 'profile' scope claims
		parser.ImplicitIdentityFieldNameName:       ValueOrNullIfEmpty(standardClaims.Name),
		parser.ImplicitIdentityFieldNameGivenName:  ValueOrNullIfEmpty(standardClaims.GivenName),
		parser.ImplicitIdentityFieldNameFamilyName: ValueOrNullIfEmpty(standardClaims.FamilyName),
		parser.ImplicitIdentityFieldNameMiddleName: ValueOrNullIfEmpty(standardClaims.MiddleName),
		parser.ImplicitIdentityFieldNameNickName:   ValueOrNullIfEmpty(standardClaims.NickName),
		parser.ImplicitIdentityFieldNameProfile:    ValueOrNullIfEmpty(standardClaims.Profile),
		parser.ImplicitIdentityFieldNamePicture:    ValueOrNullIfEmpty(standardClaims.Picture),
		parser.ImplicitIdentityFieldNameWebsite:    ValueOrNullIfEmpty(standardClaims.Website),
		parser.ImplicitIdentityFieldNameGender:     ValueOrNullIfEmpty(standardClaims.Gender),
		parser.ImplicitIdentityFieldNameZoneInfo:   ValueOrNullIfEmpty(standardClaims.ZoneInfo),
		parser.ImplicitIdentityFieldNameLocale:     ValueOrNullIfEmpty(standardClaims.Locale),
	})

	for k, v := range customClaims {
		query.AddWriteValue(Field(k), ValueOrNullIfEmpty(v))
	}

	query.AppendSelect(AllFields())
	query.AppendReturning(AllFields())

	result, err := query.UpdateStatement(ctx).ExecuteToSingle(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapToIdentity(result)
}

type ExternalUserDetails struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email-verified"`
}

func mapToIdentity(values map[string]any) (*auth.Identity, error) {
	id, ok := values["id"].(string)
	if !ok {
		return nil, errors.New("id for identity is required")
	}

	externalId, ok := values["externalId"].(string)
	if !ok {
		externalId = ""
	}

	email, ok := values["email"].(string)
	if !ok {
		email = ""
	}

	password, ok := values["password"].(string)
	if !ok {
		password = ""
	}

	issuer, ok := values["issuer"].(string)
	if !ok {
		issuer = ""
	}

	createdAt, ok := values["createdAt"].(time.Time)
	if !ok {
		return nil, errors.New("createdAt for identity is required")
	}

	updatedAt, ok := values["updatedAt"].(time.Time)
	if !ok {
		return nil, errors.New("updatedAt for identity is required")
	}

	verified, ok := values["emailVerified"].(bool)
	if !ok {
		verified = false
	}

	return &auth.Identity{
		Id:            id,
		ExternalId:    externalId,
		Email:         email,
		Password:      password,
		Issuer:        issuer,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		EmailVerified: verified,
	}, nil
}
