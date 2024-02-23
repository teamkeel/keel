package authapi

import (
	"fmt"
	"net/http"
	"strings"

	cfg "github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/jsonschema"
	"github.com/teamkeel/keel/runtime/openapi"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

func OAuthOpenApiSchema() common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx := r.Context()

		boolTrue := true
		boolFalse := false

		definition := openapi.OpenAPI{
			OpenAPI: openapi.OpenApiSpecificationVersion,
			Info: openapi.InfoObject{
				Title:   "Keel OAuth",
				Version: "1",
			},
			Paths:      map[string]openapi.PathItemObject{},
			Components: &openapi.ComponentsObject{Schemas: map[string]jsonschema.JSONSchema{}},
		}

		config, err := runtimectx.GetOAuthConfig(ctx)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		definition.Paths["/auth/providers"] = openapi.PathItemObject{
			Get: &openapi.OperationObject{
				Responses: map[string]openapi.ResponseObject{
					"200": {
						Description: "Authentication Providers",
						Content: map[string]openapi.MediaTypeObject{
							"application/json": {
								Schema: jsonschema.JSONSchema{
									Type: "array",
									Items: &jsonschema.JSONSchema{
										Ref: "#/components/schemas/ProvidersResponse",
									},
								},
							},
						},
					},
				},
			},
		}

		definition.Paths["/auth/token"] = openapi.PathItemObject{
			Post: &openapi.OperationObject{
				RequestBody: &openapi.RequestBodyObject{
					Description: "Token Request",
					Content: map[string]openapi.MediaTypeObject{
						"application/json": {
							Schema: jsonschema.JSONSchema{
								Ref: "#/components/schemas/TokenRequest",
							},
						},
						"application/x-www-form-urlencoded": {
							Schema: jsonschema.JSONSchema{
								Ref: "#/components/schemas/TokenRequest",
							},
						},
					},
					Required: &boolTrue,
				},
				Responses: map[string]openapi.ResponseObject{
					"200": {
						Description: "Token Granted",
						Content: map[string]openapi.MediaTypeObject{
							"application/json": {
								Schema: jsonschema.JSONSchema{
									Ref: "#/components/schemas/TokenResponse",
								},
							},
						},
					},
					"400": {
						Description: "Token Request Badly Formed",
						Content: map[string]openapi.MediaTypeObject{
							"application/json": {
								Schema: jsonschema.JSONSchema{
									Ref: "#/components/schemas/TokenErrorResponse",
								},
							},
						},
					},
					"401": {
						Description: "Token Request Cannot Be Granted",
						Content: map[string]openapi.MediaTypeObject{
							"application/json": {
								Schema: jsonschema.JSONSchema{
									Ref: "#/components/schemas/TokenErrorResponse",
								},
							},
						},
					},
				},
			},
		}

		for _, p := range config.Providers {
			if strings.HasPrefix(strings.ToLower(p.Name), cfg.ReservedProviderNamePrefix) {
				continue
			}

			path := fmt.Sprintf("/auth/authorize/%s", strings.ToLower(p.Name))
			definition.Paths[path] = openapi.PathItemObject{
				Get: &openapi.OperationObject{
					Responses: map[string]openapi.ResponseObject{
						"301": {
							Description: fmt.Sprintf("Single Sign-on Redirect for %s", p.Name),
						},
						"400": {
							Description: "Authorize Request Badly Formed",
							Content: map[string]openapi.MediaTypeObject{
								"application/json": {
									Schema: jsonschema.JSONSchema{
										Ref: "#/components/schemas/TokenErrorResponse",
									},
								},
							},
						},
					},
				},
			}
		}

		definition.Paths["/auth/revoke"] = openapi.PathItemObject{
			Post: &openapi.OperationObject{
				RequestBody: &openapi.RequestBodyObject{
					Description: "Token Revoke Request",
					Content: map[string]openapi.MediaTypeObject{
						"application/json": {
							Schema: jsonschema.JSONSchema{
								Ref: "#/components/schemas/RevokeRequest",
							},
						},
						"application/x-www-form-urlencoded": {
							Schema: jsonschema.JSONSchema{
								Ref: "#/components/schemas/RevokeRequest",
							},
						},
					},
					Required: &boolTrue,
				},
				Responses: map[string]openapi.ResponseObject{
					"200": {
						Description: "Token Revoked",
					},
					"400": {
						Description: "Token Revoked Request Badly Formed",
						Content: map[string]openapi.MediaTypeObject{
							"application/json": {
								Schema: jsonschema.JSONSchema{
									Ref: "#/components/schemas/TokenErrorResponse",
								},
							},
						},
					},
					"401": {
						Description: "Token Revoked Request Cannot Be Granted",
						Content: map[string]openapi.MediaTypeObject{
							"application/json": {
								Schema: jsonschema.JSONSchema{
									Ref: "#/components/schemas/TokenErrorResponse",
								},
							},
						},
					},
				},
			},
		}

		definition.Components.Schemas["ProvidersResponse"] = jsonschema.JSONSchema{
			Type: "object",
			Properties: map[string]jsonschema.JSONSchema{
				"name": {
					Type: "string",
				},
				"type": {
					Type: "string",
				},
				"authorizationUrl": {
					Type:   "integer",
					Format: "int32",
				},
				"callbackUrl": {
					Type: "string",
				},
			},
		}

		definition.Components.Schemas["TokenRequest"] = jsonschema.JSONSchema{
			UnevaluatedProperties: &boolFalse,
			OneOf: []jsonschema.JSONSchema{
				{
					Type: "object",
					Properties: map[string]jsonschema.JSONSchema{
						"grant_type": {
							Const:   "password",
							Default: "password",
						},
						"username": {
							Type: "string",
						},
						"password": {
							Type: "string",
						},
					},
					Required:             []string{"grant_type", "username", "password"},
					Title:                "Password",
					AdditionalProperties: &boolFalse,
				},
				{
					Type: "object",
					Properties: map[string]jsonschema.JSONSchema{
						"grant_type": {
							Const:   "token_exchange",
							Default: "token_exchange",
						},
						"subject_token": {
							Type: "string",
						},
					},
					Required:             []string{"grant_type", "subject_token"},
					Title:                "Token Exchange",
					AdditionalProperties: &boolFalse,
				},
				{
					Type: "object",
					Properties: map[string]jsonschema.JSONSchema{
						"grant_type": {
							Const:   "authorization_code",
							Default: "authorization_code",
						},
						"code": {
							Type: "string",
						},
					},
					Required:             []string{"grant_type", "code"},
					Title:                "Authorization Code",
					AdditionalProperties: &boolFalse,
				},
				{
					Type: "object",
					Properties: map[string]jsonschema.JSONSchema{
						"grant_type": {
							Const:   "refresh_token",
							Default: "refresh_token",
						},
						"refresh_token": {
							Type: "string",
						},
					},
					Required:             []string{"grant_type", "refresh_token"},
					Title:                "Refresh Token",
					AdditionalProperties: &boolFalse,
				},
			},
		}

		definition.Components.Schemas["RevokeRequest"] = jsonschema.JSONSchema{
			Type: "object",
			Properties: map[string]jsonschema.JSONSchema{
				"token": {
					Type: "string",
				},
			},
			Required:             []string{"token"},
			AdditionalProperties: &boolTrue,
		}

		definition.Components.Schemas["TokenResponse"] = jsonschema.JSONSchema{
			Type: "object",
			Properties: map[string]jsonschema.JSONSchema{
				"access_token": {
					Type: "string",
				},
				"token_type": {
					Type: "string",
				},
				"expires_in": {
					Type:   "integer",
					Format: "int32",
				},
				"refresh_token": {
					Type: "string",
				},
			},
		}

		definition.Components.Schemas["TokenErrorResponse"] = jsonschema.JSONSchema{
			Type: "object",
			Properties: map[string]jsonschema.JSONSchema{
				"error": {
					Type: "string",
				},
				"error_description": {
					Type: "string",
				},
			},
		}

		return common.NewJsonResponse(http.StatusOK, definition, nil)
	}
}
