package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func FindIdentityById(ctx context.Context, schema *proto.Schema, id string) (*auth.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)
	query := NewQuery(ctx, identityModel)
	err := query.Where(Field("id"), Equals, Value(id))
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
	query := NewQuery(ctx, identityModel)
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
	query := NewQuery(ctx, identityModel)
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

// Deprecated: used by the the authenticate action which is to be deprecated.
func CreateIdentity(ctx context.Context, schema *proto.Schema, email string, password string) (*auth.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(ctx, identityModel)
	query.AddWriteValues(map[string]*QueryOperand{
		"email":    Value(email),
		"password": Value(password),
		"issuer":   Value(keelIssuerClaim),
	})
	query.AppendSelect(AllFields())
	query.AppendReturning(IdField())

	result, err := query.InsertStatement().ExecuteToSingle(ctx)
	if err != nil {
		return nil, err
	}

	return mapToIdentity(result)
}

// Deprecated: used by the the authenticate action which is to be deprecated.
func CreateExternalIdentity(ctx context.Context, schema *proto.Schema, externalId string, issuer string, jwt string) (*auth.Identity, error) {
	ctx, span := tracer.Start(ctx, "Create external identity")
	defer span.End()

	span.SetAttributes(attribute.String("externalId", externalId))
	span.SetAttributes(attribute.String("issuer", issuer))

	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

	// fetch email and verified email from the openid provider if they are a known issuer
	config, err := runtimectx.GetAuthConfig(ctx)
	if err != nil {
		return nil, err
	}

	match := false
	if config != nil {
		for _, extIss := range config.Issuers {
			if issuer == extIss.Iss {
				match = true
				break
			}
		}
	}

	query := NewQuery(ctx, identityModel)
	// even if we can't fetch the user data, create it with the core information
	query.AddWriteValues(map[string]*QueryOperand{
		"externalId": Value(externalId),
		"issuer":     Value(issuer),
	})

	if match {
		externalUserDetails, err := auth.GetUserInfo(ctx, issuer, jwt)
		if err == nil {
			query.AddWriteValues(map[string]*QueryOperand{
				"email":         Value(externalUserDetails.Email),
				"emailVerified": Value(externalUserDetails.EmailVerified),
			})
		}
	}

	query.AppendSelect(AllFields())
	query.AppendReturning(IdField())

	result, err := query.InsertStatement().ExecuteToSingle(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapToIdentity(result)
}

func CreateIdentityWithIdTokenClaims(ctx context.Context, schema *proto.Schema, externalId string, issuer string, claims oauth.IdTokenClaims) (*auth.Identity, error) {
	ctx, span := tracer.Start(ctx, "Create Identity")
	defer span.End()

	span.SetAttributes(attribute.String("externalId", externalId))
	span.SetAttributes(attribute.String("issuer", issuer))

	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(ctx, identityModel)

	query.AddWriteValues(map[string]*QueryOperand{
		"externalId":    Value(externalId),
		"issuer":        Value(issuer),
		"email":         Value(claims.Email),
		"emailVerified": Value(claims.EmailVerified),
	})

	query.AppendSelect(AllFields())
	query.AppendReturning(IdField())

	result, err := query.InsertStatement().ExecuteToSingle(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return mapToIdentity(result)
}

func UpdateIdentityWithIdTokenClaims(ctx context.Context, schema *proto.Schema, externalId string, issuer string, claims oauth.IdTokenClaims) (*auth.Identity, error) {
	ctx, span := tracer.Start(ctx, "Update Identity")
	defer span.End()

	span.SetAttributes(attribute.String("externalId", claims.Subject))
	span.SetAttributes(attribute.String("issuer", claims.Issuer))

	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(ctx, identityModel)

	err := query.Where(Field("externalId"), Equals, Value(claims.Subject))
	if err != nil {
		return nil, err
	}
	query.And()
	err = query.Where(Field("issuer"), Equals, Value(claims.Issuer))
	if err != nil {
		return nil, err
	}

	query.AddWriteValues(map[string]*QueryOperand{
		"email":         Value(claims.Email),
		"emailVerified": Value(claims.EmailVerified),
	})

	query.AppendSelect(AllFields())
	query.AppendReturning(AllFields())

	result, err := query.UpdateStatement().ExecuteToSingle(ctx)
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
	if _, err := ksuid.Parse(id); err != nil {
		return nil, fmt.Errorf("id for identity cannot be parsed: %s", values["id"])
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
