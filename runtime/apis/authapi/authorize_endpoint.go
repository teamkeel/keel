package authapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dchest/uniuri"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
	"golang.org/x/oauth2"
)

// Error response types for the authorization endpoint
// https://datatracker.ietf.org/doc/html/rfc6749#section-4.1.2.1
const (
	// The request is missing a required parameter, includes an
	// invalid parameter value, includes a parameter more than
	// once, or is otherwise malformed.
	AuthorizationErrInvalidRequest = "invalid_request"
	// The client is not authorized to request an authorization
	// code using this method.
	AuthorizationErrUnauthorizedClient = "unauthorized_client"
	// The resource owner or authorization server denied the
	// request.
	AuthorizationErrAccessDenied = "access_denied"
	// The authorization server encountered an unexpected
	// condition that prevented it from fulfilling the request.
	// (This error code is needed because a 500 Internal Server
	// Error HTTP status code cannot be returned to the client
	// via an HTTP redirect.)
	AuthorizationErrServerError = "server_error"
)

// AuthorizeHandler is a redirection endpoint that will redirect to the provider's sign-in/auth page
func AuthorizeHandler(schema *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "Authorize Endpoint")
		defer span.End()

		provider, err := providerFromPath(ctx, r.URL)
		if err != nil {
			return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, "login url malformed or provider not found", err)
		}

		config, err := runtimectx.GetOAuthConfig(ctx)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		if config.RedirectUrl == nil {
			return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, "redirectUrl must be specified in keelconfig.yaml", err)
		}

		// If the secret is not yet, then package this up and send it as an error with the redirectUrl
		_, hasSecret := GetClientSecret(ctx, provider)
		if !hasSecret {
			err := fmt.Errorf("client secret not configured for provider: %s", provider.Name)
			return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, err.Error(), err)
		}

		issuer, hasIssuer := provider.GetIssuerUrl()
		if !hasIssuer {
			return common.InternalServerErrorResponse(ctx, err)
		}

		// Establishes new OIDC provider. This will call the providers discovery endpoint
		oidcProvider, err := oidc.NewProvider(ctx, issuer)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		callbackUrl, err := provider.GetCallbackUrl()
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		// RedirectURL is needed for provider to redirect back to Keel
		// Secret is _not_ required when getting auth code
		oauthConfig := &oauth2.Config{
			ClientID: provider.ClientId,
			Endpoint: oauth2.Endpoint{
				AuthURL:  oidcProvider.Endpoint().AuthURL,
				TokenURL: oidcProvider.Endpoint().TokenURL,
			},
			Scopes:      []string{"openid", "email", "profile"},
			RedirectURL: callbackUrl.String(),
		}

		u := oauthConfig.AuthCodeURL(uniuri.New())

		redirectUrl, err := url.Parse(u)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		return common.NewRedirectResponse(redirectUrl)
	}
}

// CallbackHandler is called by the provider after the authentication process is complete
//
// If there is something wrong with the syntax of the request, such as the redirect_uri or client_id is invalid,
// then it’s important not to redirect the user and instead you should show the error message directly.
// This is to avoid letting your authorization server be used as an open redirector.
func CallbackHandler(schema *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "Callback Endpoint")
		defer span.End()

		provider, err := providerFromPath(ctx, r.URL)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		issuer, hasIssuer := provider.GetIssuerUrl()
		if !hasIssuer {
			return common.InternalServerErrorResponse(ctx, err)
		}

		cfg, err := runtimectx.GetOAuthConfig(ctx)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		if cfg.RedirectUrl == nil {
			return common.InternalServerErrorResponse(ctx, fmt.Errorf("redirectUrl not set"))
		}

		redirectUrl, err := url.Parse(*cfg.RedirectUrl)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		// If the auth provider errored, then package this up and send it as an error with the redirectUrl
		if callbackError := r.URL.Query().Get("error"); callbackError != "" {
			err := fmt.Errorf("provider error: %s. %s", callbackError, r.URL.Query().Get("error_description"))
			return redirectErrResponse(ctx, redirectUrl, AuthorizationErrAccessDenied, err.Error(), err)
		}

		oidcProv, err := oidc.NewProvider(ctx, issuer)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		secret, hasSecret := GetClientSecret(ctx, provider)
		if !hasSecret {
			return common.InternalServerErrorResponse(ctx, err)
		}

		callbackUrl, err := provider.GetCallbackUrl()
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		// ClientSecret is required for token exchange
		oauthConfig := &oauth2.Config{
			ClientID:     provider.ClientId,
			ClientSecret: secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  oidcProv.Endpoint().AuthURL,
				TokenURL: oidcProv.Endpoint().TokenURL,
			},
			RedirectURL: callbackUrl.String(),
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			return common.InternalServerErrorResponse(ctx, errors.New("code not returned with callback url"))
		}

		// If the token exchange fails, then package this up and send it as an error with the redirectUrl
		token, err := oauthConfig.Exchange(ctx, code)
		if err != nil {
			var receiveErr *oauth2.RetrieveError
			if errors.As(err, &receiveErr) && receiveErr.ErrorCode != "" {
				return redirectErrResponse(ctx, redirectUrl, receiveErr.ErrorCode, receiveErr.ErrorDescription, receiveErr)
			}

			return redirectErrResponse(ctx, redirectUrl, AuthorizationErrAccessDenied, "failed to exchange code at provider token endpoint", err)
		}

		if !token.Valid() {
			err := errors.New("access token is not valid or has expired")
			return redirectErrResponse(ctx, redirectUrl, AuthorizationErrAccessDenied, err.Error(), err)
		}

		// Extract the ID Token from the OAuth2 request.
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			err := errors.New("provider did not respond with an id token")
			return redirectErrResponse(ctx, redirectUrl, AuthorizationErrServerError, err.Error(), err)
		}

		var verifier = oidcProv.Verifier(&oidc.Config{
			ClientID: provider.ClientId,
		})

		// Verify the ID token with the OIDC provider
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			return redirectErrResponse(ctx, redirectUrl, AuthorizationErrAccessDenied, "falied to verify ID token with OIDC provider", err)
		}

		// Extract standardClaims
		var standardClaims oauth.IdTokenClaims
		if err := idToken.Claims(&standardClaims); err != nil {
			return redirectErrResponse(ctx, redirectUrl, AuthorizationErrServerError, "insufficient claims on id_token", err)
		}

		var claims map[string]any
		if err := idToken.Claims(&claims); err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		customClaims := map[string]any{}
		for _, c := range cfg.Claims {
			customClaims[c.Field] = claims[c.Key]
		}

		var identity auth.Identity
		identity, err = actions.FindIdentityByExternalId(ctx, schema, idToken.Subject, idToken.Issuer)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		if identity == nil {
			identity, err = actions.CreateIdentityWithClaims(ctx, schema, idToken.Subject, idToken.Issuer, &standardClaims, customClaims)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}

			ctx = auth.WithIdentity(ctx, identity)
			err = functions.CallPredefinedHook(ctx, config.HookAfterIdentityCreated)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}
		} else {
			identity, err = actions.UpdateIdentityWithClaims(ctx, schema, idToken.Subject, idToken.Issuer, &standardClaims, customClaims)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}
		}

		authCode, err := oauth.NewAuthCode(ctx, identity[parser.FieldNameId].(string))
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		values := url.Values{}
		values.Add("code", authCode)
		redirectUrl.RawQuery = values.Encode()

		return common.NewRedirectResponse(redirectUrl)
	}
}

func GetClientSecret(ctx context.Context, provider *config.Provider) (string, bool) {
	name := provider.GetClientSecretName()
	secret, err := runtimectx.GetSecret(ctx, name)
	return secret, err == nil
}

func providerFromPath(ctx context.Context, url *url.URL) (*config.Provider, error) {
	config, err := runtimectx.GetOAuthConfig(ctx)
	if err != nil {
		return nil, err
	}

	p := strings.Split(strings.Trim(url.Path, "/"), "/")

	if len(p) != 3 {
		return nil, fmt.Errorf("invalid login path: %s", url.Path)
	}

	provider := config.GetProvider(p[2])
	if provider == nil {
		return nil, fmt.Errorf("no provider with the name '%s' has been configured", p[2])
	}

	return provider, nil
}
