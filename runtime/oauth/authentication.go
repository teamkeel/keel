package oauth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime")

// VerifyIdToken will verify the ID token from an OpenID Connect provider.
func VerifyIdToken(ctx context.Context, idTokenRaw string) (*oidc.IDToken, error) {
	ctx, span := tracer.Start(ctx, "Verify ID Token")
	defer span.End()

	issuer, err := auth.ExtractClaimFromToken(idTokenRaw, "iss")
	if issuer == "" {
		return nil, err
	}
	span.AddEvent("Issuer extracted from ID Token")

	span.SetAttributes(attribute.String("issuer", issuer))

	authConfig, err := runtimectx.GetAuthConfig(ctx)
	if err != nil {
		return nil, err
	}

	issuerPermitted := authConfig.AllowAnyIssuers
	if !issuerPermitted {
		for _, e := range authConfig.Issuers {
			if e.Iss == issuer {
				issuerPermitted = true
			}
		}
	}

	if !issuerPermitted {
		return nil, fmt.Errorf("issuer %s not registered to authenticate on this server", issuer)
	}

	// Establishes new OIDC provider. This will call the providers discovery endpoint.
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	span.AddEvent("Provider's ODIC config fetched")

	// TODO: what are we missing by skipping the client ID check?
	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	// Verify that the ID token legitimately was signed by the provider and that it has not expired.
	idToken, err := verifier.Verify(ctx, idTokenRaw)
	if err != nil {
		return nil, err
	}
	span.AddEvent("ID Token verified")

	return idToken, nil
}
