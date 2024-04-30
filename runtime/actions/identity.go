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

	return result, nil
}

func FindIdentityByEmail(ctx context.Context, schema *proto.Schema, email string, issuer string) (auth.Identity, error) {
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

	return result, nil
}

func FindIdentityByExternalId(ctx context.Context, schema *proto.Schema, externalId string, issuer string) (auth.Identity, error) {
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

	return result, nil
}

func CreateIdentity(ctx context.Context, schema *proto.Schema, email string, password string, issuer string) (auth.Identity, error) {
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

	return result, nil
}

func CreateIdentityWithClaims(ctx context.Context, schema *proto.Schema, externalId string, issuer string, standardClaims *oauth.IdTokenClaims, customClaims map[string]any) (auth.Identity, error) {
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

	return result, nil
}

func UpdateIdentityWithClaims(ctx context.Context, schema *proto.Schema, externalId string, issuer string, standardClaims *oauth.IdTokenClaims, customClaims map[string]any) (auth.Identity, error) {
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

	return result, nil
}
