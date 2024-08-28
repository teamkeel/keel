package actions

import (
	"context"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func FindIdentityById(ctx context.Context, schema *proto.Schema, id string) (auth.Identity, error) {
	identityModel := schema.FindModel(parser.IdentityModelName)
	query := NewQuery(identityModel)
	err := query.Where(IdField(), Equals, Value(id))
	if err != nil {
		return nil, err
	}

	query.Select(AllFields())
	result, err := query.SelectStatement().ExecuteToSingle(ctx)

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return result, nil
}

func FindIdentityByEmail(ctx context.Context, schema *proto.Schema, email string, issuer string) (auth.Identity, error) {
	identityModel := schema.FindModel(parser.IdentityModelName)
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

	query.Select(AllFields())
	result, err := query.SelectStatement().ExecuteToSingle(ctx)

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return result, nil
}

func FindIdentityByExternalId(ctx context.Context, schema *proto.Schema, externalId string, issuer string) (auth.Identity, error) {
	identityModel := schema.FindModel(parser.IdentityModelName)
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

	query.Select(AllFields())
	result, err := query.SelectStatement().ExecuteToSingle(ctx)

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return result, nil
}

func CreateIdentity(ctx context.Context, schema *proto.Schema, email string, password string, issuer string) (auth.Identity, error) {
	identityModel := schema.FindModel(parser.IdentityModelName)

	query := NewQuery(identityModel)
	query.AddWriteValues(map[string]*QueryOperand{
		"email":    Value(email),
		"password": Value(password),
		"issuer":   Value(issuer),
	})
	query.Select(AllFields())
	query.AppendReturning(IdField())

	result, err := query.InsertStatement(ctx).ExecuteToSingle(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func CreateIdentityWithClaims(ctx context.Context, schema *proto.Schema, externalId string, issuer string, standardClaims *oauth.IdTokenClaims, customClaims map[string]any) (auth.Identity, error) {
	ctx, span := tracer.Start(ctx, "Create Identity")
	defer span.End()

	span.SetAttributes(attribute.String("externalId", externalId))
	span.SetAttributes(attribute.String("issuer", issuer))

	identityModel := schema.FindModel(parser.IdentityModelName)

	query := NewQuery(identityModel)

	query.AddWriteValues(map[string]*QueryOperand{
		// default 'email' scope claims
		parser.IdentityFieldNameExternalId: Value(externalId),
		parser.IdentityFieldNameIssuer:     Value(issuer),

		// default 'profile' scope claims
		parser.IdentityFieldNameEmail:         Value(standardClaims.Email),
		parser.IdentityFieldNameEmailVerified: Value(standardClaims.EmailVerified),

		// default 'profile' scope claims
		parser.IdentityFieldNameName:       ValueOrNullIfEmpty(standardClaims.Name),
		parser.IdentityFieldNameGivenName:  ValueOrNullIfEmpty(standardClaims.GivenName),
		parser.IdentityFieldNameFamilyName: ValueOrNullIfEmpty(standardClaims.FamilyName),
		parser.IdentityFieldNameMiddleName: ValueOrNullIfEmpty(standardClaims.MiddleName),
		parser.IdentityFieldNameNickName:   ValueOrNullIfEmpty(standardClaims.NickName),
		parser.IdentityFieldNameProfile:    ValueOrNullIfEmpty(standardClaims.Profile),
		parser.IdentityFieldNamePicture:    ValueOrNullIfEmpty(standardClaims.Picture),
		parser.IdentityFieldNameWebsite:    ValueOrNullIfEmpty(standardClaims.Website),
		parser.IdentityFieldNameGender:     ValueOrNullIfEmpty(standardClaims.Gender),
		parser.IdentityFieldNameZoneInfo:   ValueOrNullIfEmpty(standardClaims.ZoneInfo),
		parser.IdentityFieldNameLocale:     ValueOrNullIfEmpty(standardClaims.Locale),
	})

	for k, v := range customClaims {
		query.AddWriteValue(Field(k), ValueOrNullIfEmpty(v))
	}

	query.Select(AllFields())
	query.AppendReturning(IdField())

	result, err := query.InsertStatement(ctx).ExecuteToSingle(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return result, nil
}

func UpdateIdentityWithClaims(ctx context.Context, schema *proto.Schema, externalId string, issuer string, standardClaims *oauth.IdTokenClaims, customClaims map[string]any) (auth.Identity, error) {
	ctx, span := tracer.Start(ctx, "Update Identity")
	defer span.End()

	span.SetAttributes(attribute.String("externalId", standardClaims.Subject))
	span.SetAttributes(attribute.String("issuer", standardClaims.Issuer))

	identityModel := schema.FindModel(parser.IdentityModelName)

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
		parser.IdentityFieldNameEmail:         Value(standardClaims.Email),
		parser.IdentityFieldNameEmailVerified: Value(standardClaims.EmailVerified),

		// default 'profile' scope claims
		parser.IdentityFieldNameName:       ValueOrNullIfEmpty(standardClaims.Name),
		parser.IdentityFieldNameGivenName:  ValueOrNullIfEmpty(standardClaims.GivenName),
		parser.IdentityFieldNameFamilyName: ValueOrNullIfEmpty(standardClaims.FamilyName),
		parser.IdentityFieldNameMiddleName: ValueOrNullIfEmpty(standardClaims.MiddleName),
		parser.IdentityFieldNameNickName:   ValueOrNullIfEmpty(standardClaims.NickName),
		parser.IdentityFieldNameProfile:    ValueOrNullIfEmpty(standardClaims.Profile),
		parser.IdentityFieldNamePicture:    ValueOrNullIfEmpty(standardClaims.Picture),
		parser.IdentityFieldNameWebsite:    ValueOrNullIfEmpty(standardClaims.Website),
		parser.IdentityFieldNameGender:     ValueOrNullIfEmpty(standardClaims.Gender),
		parser.IdentityFieldNameZoneInfo:   ValueOrNullIfEmpty(standardClaims.ZoneInfo),
		parser.IdentityFieldNameLocale:     ValueOrNullIfEmpty(standardClaims.Locale),
	})

	for k, v := range customClaims {
		query.AddWriteValue(Field(k), ValueOrNullIfEmpty(v))
	}

	query.Select(AllFields())
	query.AppendReturning(AllFields())

	result, err := query.UpdateStatement(ctx).ExecuteToSingle(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return result, nil
}
