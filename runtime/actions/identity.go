package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func FindIdentityById(ctx context.Context, schema *proto.Schema, id string) (*runtimectx.Identity, error) {
	return findSingle(ctx, schema, "id", id)
}

func FindIdentityByEmail(ctx context.Context, schema *proto.Schema, email string) (*runtimectx.Identity, error) {
	return findSingle(ctx, schema, "email", email)
}

func FindIdentityByExternalId(ctx context.Context, schema *proto.Schema, externalId string, issuer string) (*runtimectx.Identity, error) {
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

func findSingle(ctx context.Context, schema *proto.Schema, field string, value string) (*runtimectx.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)
	query := NewQuery(identityModel)
	err := query.Where(Field(field), Equals, Value(value))
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

func CreateIdentity(ctx context.Context, schema *proto.Schema, email string, password string) (*runtimectx.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(identityModel)
	query.AddWriteValues(map[string]any{
		"email":    email,
		"password": password,
		"issuer":   keelIssuerClaim,
	})
	query.AppendSelect(AllFields())
	query.AppendReturning(IdField())

	result, err := query.InsertStatement().ExecuteToSingle(ctx)
	if err != nil {
		return nil, err
	}

	return mapToIdentity(result)
}

func CreateExternalIdentity(ctx context.Context, schema *proto.Schema, externalId string, issuer string, jwt string) (*runtimectx.Identity, error) {
	ctx, span := tracer.Start(ctx, "Create external identity")
	defer span.End()

	span.SetAttributes(attribute.String("externalId", externalId))
	span.SetAttributes(attribute.String("issuer", issuer))

	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

	// fetch email and verified email from the openid provider if they are a known issuer
	config, _ := runtimectx.GetAuthConfig(ctx)

	match := false
	if config != nil {
		for _, extIss := range config.Issuers {
			if issuer == extIss.Iss {
				match = true
				break
			}
		}
	}

	query := NewQuery(identityModel)
	// even if we can't fetch the user data, create it with the core information
	query.AddWriteValues(map[string]any{
		"externalId": externalId,
		"issuer":     issuer,
	})

	if match {
		externalUserDetails, err := auth.GetUserInfo(ctx, issuer, jwt)
		if err == nil {
			query.AddWriteValues(map[string]any{
				"email":         externalUserDetails.Email,
				"emailVerified": externalUserDetails.EmailVerified,
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

type ExternalUserDetails struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email-verified"`
}

func mapToIdentity(values map[string]any) (*runtimectx.Identity, error) {
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

	return &runtimectx.Identity{
		Id:         id,
		ExternalId: externalId,
		Email:      email,
		Password:   password,
		Issuer:     issuer,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}
