package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
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
	err = query.Where(Field("createdBy"), Equals, Value(issuer))
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
		"email":     email,
		"password":  password,
		"createdBy": keelIssuerClaim,
	})
	query.AppendSelect(AllFields())
	query.AppendReturning(IdField())

	result, err := query.InsertStatement().ExecuteToSingle(ctx)
	if err != nil {
		return nil, err
	}

	return mapToIdentity(result)
}

func CreateExternalIdentity(ctx context.Context, schema *proto.Schema, externalId string, createdBy string) (*runtimectx.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(identityModel)
	query.AddWriteValues(map[string]any{
		"externalId": externalId,
		"createdBy":  createdBy,
	})
	query.AppendSelect(AllFields())
	query.AppendReturning(IdField())

	result, err := query.InsertStatement().ExecuteToSingle(ctx)
	if err != nil {
		return nil, err
	}

	return mapToIdentity(result)
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

	createdBy, ok := values["createdBy"].(string)
	if !ok {
		createdBy = ""
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
		CreatedBy:  createdBy,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}
