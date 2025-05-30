package oauth

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type IdTokenClaims struct {
	jwt.RegisteredClaims
	UserClaims
}

// https://openid.net/specs/openid-connect-basic-1_0.html#StandardClaims

type UserClaims struct {
	// default 'email' scope claims
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`

	// default 'profile' scope claims
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	MiddleName        string `json:"middle_name,omitempty"`
	NickName          string `json:"nick_name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Profile           string `json:"profile,omitempty"`
	Picture           string `json:"picture,omitempty"`
	Website           string `json:"website,omitempty"`
	Gender            string `json:"gender,omitempty"`
	ZoneInfo          string `json:"zoneinfo,omitempty"`
	Locale            string `json:"locale,omitempty"`

	// default 'phone' scope claims
	PhoneNumber         string `json:"phone_number,omitempty"`
	PhoneNumberVerified bool   `json:"phone_number_verified,omitempty"`
}

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime")

// VerifyIdToken will verify the ID token from an OpenID Connect provider.
func VerifyIdToken(ctx context.Context, idTokenRaw string) (*oidc.IDToken, error) {
	ctx, span := tracer.Start(ctx, "Verify ID Token")
	defer span.End()

	issuer, err := ExtractClaimFromJwt(idTokenRaw, "iss")
	if err != nil {
		return nil, err
	}
	if issuer == "" {
		return nil, errors.New("iss claim cannot be an empty string")
	}

	span.AddEvent("Issuer extracted from ID Token")
	span.SetAttributes(attribute.String("issuer", issuer))

	authConfig, err := runtimectx.GetOAuthConfig(ctx)
	if err != nil {
		return nil, err
	}

	providers, err := authConfig.GetOidcProvidersByIssuer(issuer)
	if err != nil {
		return nil, err
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("issuer %s not registered to authenticate on this server", issuer)
	}

	// Establishes new OIDC provider. This will call the providers discovery endpoint
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	span.AddEvent("Provider's ODIC config fetched")

	// Verify against each configured provider with this issuer
	var verificationErrs error
	for _, p := range providers {
		// Checking the clientId during verification ensures that the ID token was intended for this client,
		// because it could have been stolen from any other application with an ID token from this same issuer.
		oidcConfig := &oidc.Config{
			ClientID: p.ClientId,
		}

		// Verify that the ID token legitimately was signed by the provider and that it has not expired
		idToken, err := provider.Verifier(oidcConfig).Verify(ctx, idTokenRaw)
		if err != nil {
			verificationErrs = errors.Join(err, verificationErrs)
		} else {
			span.AddEvent("ID Token verified")
			return idToken, nil
		}
	}

	return nil, verificationErrs
}
