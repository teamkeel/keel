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

// LoginHandler will redirect to the specified provider in order to authenticate the user
func LoginHandler(schema *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "Login Endpoint")
		defer span.End()

		provider, err := providerFromPath(ctx, r.URL)
		if err != nil {
			return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, "login url malformed or provider not found", err)
		}

		secret, hasSecret := provider.GetClientSecret()
		if !hasSecret {
			err = fmt.Errorf("client secret not configured for provider: %s", provider.Name)
			return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, err.Error(), err)
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

		issuer, hasIssuer := provider.GetIssuer()
		if !hasIssuer {
			return common.InternalServerErrorResponse(ctx, err)
		}

		config, err := runtimectx.GetOAuthConfig(ctx)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		if config.RedirectUrl == nil {
			return common.InternalServerErrorResponse(ctx, fmt.Errorf("redirectUrl not set"))
		}

		redirectUrl, err := url.Parse(*config.RedirectUrl)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		// If the auth provider errored, then package this up and send it as an error with the redirectUrl
		if callbackError := r.FormValue("error"); callbackError != "" {
			err := fmt.Errorf("provider error: %s. %s", callbackError, r.FormValue("error_description"))
			return redirectErrResponse(ctx, redirectUrl, AuthorizationErrAccessDenied, err.Error(), err)
		}

		oidcProv, err := oidc.NewProvider(ctx, issuer)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		// If the secret is not yet, then package this up and send it as an error with the redirectUrl
		secret, hasSecret := provider.GetClientSecret()
		if !hasSecret {
			err := fmt.Errorf("client secret not configured for provider: %s", provider.Name)
			return redirectErrResponse(ctx, redirectUrl, AuthorizationErrAccessDenied, err.Error(), err)
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

		// If the token exchange fails, then package this up and send it as an error with the redirectUrl
		token, err := oauthConfig.Exchange(ctx, code)
		if err != nil {
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

		// Extract claims
		var claims oauth.IdTokenClaims
		if err := idToken.Claims(&claims); err != nil {
			return redirectErrResponse(ctx, redirectUrl, AuthorizationErrServerError, "insufficient claims on id_token", err)
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

		authCode, err := oauth.NewAuthCode(ctx, identity.Id)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		values := url.Values{}
		values.Add("code", authCode)
		redirectUrl.RawQuery = values.Encode()

		return common.NewRedirectResponse(redirectUrl)
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
