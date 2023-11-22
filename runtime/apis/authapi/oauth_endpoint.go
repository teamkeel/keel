package authapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/dchest/uniuri"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
)

func OAuthHandler(schema *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		return common.Response{
			Status: http.StatusNotImplemented,
		}
	}
}

// LoginHandler will redirect to the specified provider in order to authenticate with the user
func LoginHandler(schema *proto.Schema) func(http.ResponseWriter, *http.Request) common.Response {
	return func(w http.ResponseWriter, r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "Login Endpoint")
		defer span.End()

		provider, err := providerFromPath(ctx, r.URL)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			return authErrResponse(ctx, http.StatusBadRequest, InvalidRequest, "login url malformed or provider not found")
		}

		secret, hasSecret := provider.GetClientSecret()
		if !hasSecret {
			err = fmt.Errorf("client secret not configured for provider: %s", provider.Name)
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			return authErrResponse(ctx, http.StatusBadRequest, InvalidRequest, err.Error())
		}

		issuer, hasIssuer := provider.GetIssuer()
		if !hasIssuer {
			err = fmt.Errorf("no issuer available for sso login with provider: %s", provider.Name)
			return common.InternalServerErrorResponse(ctx, err)
		}

		oidcProv, err := oidc.NewProvider(ctx, issuer)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		oauthConfig := &oauth2.Config{
			ClientID:     provider.ClientId,
			ClientSecret: secret,
			Endpoint:     oidcProv.Endpoint(),
			Scopes:       []string{"openid", "email", "profile"},
			RedirectURL:  "http://" + r.Host + "/auth/callback/" + strings.ToLower(provider.Name),
		}

		url := oauthConfig.AuthCodeURL(uniuri.New())
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)

		return common.NewJsonResponse(http.StatusTemporaryRedirect, "login handler redirect", nil)
	}
}

// CallbackHandler is called by the provider after the authentication process is complete
func CallbackHandler(schema *proto.Schema) func(http.ResponseWriter, *http.Request) common.Response {
	return func(w http.ResponseWriter, r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "Callback Endpoint")
		defer span.End()

		if callbackError := r.FormValue("error"); callbackError != "" {
			err := fmt.Errorf("provider could not authenticate due to %s: %s", callbackError, r.FormValue("error_description"))
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			return authErrResponse(ctx, http.StatusBadRequest, InvalidRequest, err.Error())
		}

		provider, err := providerFromPath(ctx, r.URL)
		if err != nil {
			err = fmt.Errorf("no issuer available for sso login with provider: %s", provider.Name)
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			return authErrResponse(ctx, http.StatusBadRequest, InvalidRequest, "callback url malformed or provider not found")
		}

		issuer, hasIssuer := provider.GetIssuer()
		if !hasIssuer {
			err = fmt.Errorf("no issuer available for sso login with provider: %s", provider.Name)
			return common.InternalServerErrorResponse(ctx, err)
		}

		oidcProv, err := oidc.NewProvider(ctx, issuer)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		secret, hasSecret := provider.GetClientSecret()
		if !hasSecret {
			err := fmt.Errorf("client secret not configured for provider: %s", provider.Name)
			return common.InternalServerErrorResponse(ctx, err)
		}

		oauthConfig := &oauth2.Config{
			ClientID:     provider.ClientId,
			ClientSecret: secret,
			Endpoint:     oidcProv.Endpoint(),
			RedirectURL:  "http://" + r.Host + "/auth/callback/" + strings.ToLower(provider.Name),
		}

		code := r.FormValue("code")
		if !r.Form.Has("code") || code == "" {
			return common.InternalServerErrorResponse(ctx, errors.New("code not returned with callback url"))
		}

		token, err := oauthConfig.Exchange(ctx, code)
		if err != nil {
			// todo parse oauth error message from providers token endpoint
			return common.InternalServerErrorResponse(ctx, err)
		}

		if !token.Valid() {
			return authErrResponse(ctx, http.StatusUnauthorized, InvalidClient, "access token is not valid or has expired")
		}

		// Extract the ID Token from the OAuth2 request.
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			return authErrResponse(ctx, http.StatusBadRequest, InvalidRequest, "provider did not respond with an id token")
		}

		var verifier = oidcProv.Verifier(&oidc.Config{
			ClientID: provider.ClientId,
		})

		// Verify the ID token with the OIDC provider
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		// Extract claims
		var claims oauth.IdTokenClaims
		if err := idToken.Claims(&claims); err != nil {
			return authErrResponse(ctx, http.StatusBadRequest, InvalidRequest, "insufficient claims on id_token")
		}

		var identity *auth.Identity
		identity, err = actions.FindIdentityByExternalId(ctx, schema, idToken.Subject, idToken.Issuer)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		if identity == nil {
			identity, err = actions.CreateIdentityWithIdTokenClaims(ctx, schema, idToken.Subject, idToken.Issuer, claims)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}
		} else {
			identity, err = actions.UpdateIdentityWithIdTokenClaims(ctx, schema, idToken.Subject, idToken.Issuer, claims)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}
		}

		config, err := runtimectx.GetOAuthConfig(ctx)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		if config.RedirectUrl == nil {
			err := fmt.Errorf("redirectUrl not set")
			return common.InternalServerErrorResponse(ctx, err)
		}

		authCode, err := oauth.NewAuthCode(ctx, identity.Id)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		redirectUrl, err := url.Parse(*config.RedirectUrl)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		values := url.Values{}
		values.Add("code", authCode)
		redirectUrl.RawQuery = values.Encode()

		http.Redirect(w, r, redirectUrl.String(), http.StatusFound)

		return common.NewJsonResponse(http.StatusFound, "callback handler redirect", nil)
	}

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

func authErrResponse(ctx context.Context, status int, errorType string, errorDescription string) common.Response {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Error, errorType)

	span.SetAttributes(
		attribute.String("auth.error", errorType),
		attribute.String("auth.error_description", errorDescription),
	)

	return common.NewJsonResponse(status, &ErrorResponse{
		Error:            errorType,
		ErrorDescription: errorDescription,
	}, nil)
}
