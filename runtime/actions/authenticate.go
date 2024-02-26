package actions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	email "net/mail"
	"net/url"
	"strings"

	"github.com/karlseguin/typed"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/teamkeel/keel/mail"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"

	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken     = common.NewAuthenticationFailedMessageErr("cannot be parsed or verified as a valid JWT")
	ErrTokenExpired     = common.NewAuthenticationFailedMessageErr("token has expired")
	ErrIdentityNotFound = common.NewAuthenticationFailedMessageErr("identity not found")
)

func ResetRequestPassword(scope *Scope, input map[string]any) error {
	var err error
	typedInput := typed.New(input)

	emailString := typedInput.String("email")
	if _, err = email.ParseAddress(emailString); err != nil {
		return common.RuntimeError{Code: common.ErrInvalidInput, Message: "invalid email address"}
	}

	var redirectUrl *url.URL
	if redirectUrl, err = url.ParseRequestURI(typedInput.String("redirectUrl")); err != nil {
		return common.RuntimeError{Code: common.ErrInvalidInput, Message: "invalid redirect URL"}
	}

	var identity *auth.Identity
	identity, err = FindIdentityByEmail(scope.Context, scope.Schema, emailString, oauth.KeelIssuer)
	if err != nil {
		return err
	}
	if identity == nil {
		return nil
	}

	token, err := oauth.GenerateResetToken(scope.Context, identity.Id)
	if err != nil {
		return err
	}

	q := redirectUrl.Query()
	q.Add("token", token)
	redirectUrl.RawQuery = q.Encode()

	client, err := runtimectx.GetMailClient(scope.Context)
	if err != nil {
		return err
	}

	err = client.Send(scope.Context, &mail.SendEmailRequest{
		To:        identity.Email,
		From:      "hi@keel.xyz",
		Subject:   "[Keel] Reset password request",
		PlainText: fmt.Sprintf("Please follow this link to reset your password: %s", redirectUrl),
	})

	return err
}

// Deprecated: we will be deprecating the authenticate action and password flow in favour of the new auth endpoints
func ResetPassword(scope *Scope, input map[string]any) error {
	typedInput := typed.New(input)

	token := typedInput.String("token")
	password := typedInput.String("password")

	identityId, err := oauth.ValidateResetToken(scope.Context, token)
	switch {
	case errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenExpired):
		return common.RuntimeError{Code: common.ErrInvalidInput, Message: err.Error()}
	case err != nil:
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	identityModel := proto.FindModel(scope.Schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(identityModel)
	err = query.Where(IdField(), Equals, Value(identityId))
	if err != nil {
		return err
	}

	query.AddWriteValue(Field("password"), Value(string(hashedPassword)))

	affected, err := query.UpdateStatement(scope.Context).Execute(scope.Context)
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("expected 1 row to be updated, but %v rows were updated", affected)
	}

	return nil
}

func HandleAuthorizationHeader(ctx context.Context, schema *proto.Schema, headers http.Header) (*auth.Identity, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return nil, nil
	}

	headerSplit := strings.Split(header, "Bearer ")
	if len(headerSplit) != 2 {
		return nil, common.NewAuthenticationFailedMessageErr("no 'Bearer' prefix in the Authorization header")
	}

	token := headerSplit[1]

	if token != "" {
		identity, err := HandleBearerToken(ctx, schema, token)
		if err != nil {
			return nil, err
		}
		return identity, nil
	}

	return nil, nil
}

func HandleBearerToken(ctx context.Context, schema *proto.Schema, token string) (*auth.Identity, error) {
	ctx, span := tracer.Start(ctx, "Authorization")
	defer span.End()

	subject, err := oauth.ValidateAccessToken(ctx, token)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	identity, err := FindIdentityById(ctx, schema, subject)
	if err != nil {
		return nil, err
	}

	if identity == nil {
		return nil, ErrIdentityNotFound
	}

	span.SetAttributes(attribute.String("identity.id", identity.Id))

	return identity, nil
}
