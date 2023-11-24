package authapi

import (
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

type ProviderResponse struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	AuthorizeUrl string `json:"authorizeUrl"`
	CallbackUrl  string `json:"callbackUrl"`
}

func ProvidersHandler(schema *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "Providers")
		defer span.End()

		config, err := runtimectx.GetOAuthConfig(ctx)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		providers := []ProviderResponse{}

		for _, p := range config.Providers {
			authUrl, err := p.GetAuthorizeUrl()
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}

			callbackUrl, err := p.GetCallbackUrl()
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}

			providers = append(providers, ProviderResponse{
				Type:         p.Type,
				Name:         p.Name,
				AuthorizeUrl: authUrl.String(),
				CallbackUrl:  callbackUrl.String(),
			})
		}

		return common.NewJsonResponse(http.StatusOK, providers, nil)
	}
}
