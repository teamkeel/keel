package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/rs/cors"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

type Request struct {
	Context context.Context
	Path    string
	Body    []byte
}

type Response struct {
	Body   []byte
	Status int
}

type Handler func(r *Request) (*Response, error)

func Serve(currSchema *proto.Schema) func(w http.ResponseWriter, r *http.Request) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if currSchema == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Cannot serve requests when schema contains errors"))
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		handler := NewHandler(currSchema)

		identityId, err := RetrieveIdentityClaim(r)

		switch {
		case errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenExpired):
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Valid bearer token required to authenticate"))
			return
		case errors.Is(err, ErrNoBearerPrefix):
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		case errors.Is(err, ErrInvalidIdentityClaim):
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := r.Context()
		ctx = runtimectx.WithIdentity(ctx, identityId)

		response, err := handler(&Request{
			Context: ctx,
			Path:    r.URL.Path,
			Body:    body,
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(response.Status)
		w.Write(response.Body)
	}

	handler := http.HandlerFunc(h)
	return cors.Default().Handler(handler).ServeHTTP

}

func NewHandler(s *proto.Schema) Handler {
	handlers := map[string]Handler{}

	for _, api := range s.Apis {
		switch api.Type {
		case proto.ApiType_API_TYPE_GRAPHQL:
			handlers["/"+api.Name] = NewGraphQLHandler(s, api)
		default:
			panic(fmt.Sprintf("api type %s not supported", api.Type.String()))
		}
	}

	return func(r *Request) (*Response, error) {
		path := strings.TrimSuffix(r.Path, "/")

		handler, ok := handlers[path]
		if !ok {
			return &Response{
				Status: 404,
				Body:   []byte("Not found"),
			}, nil
		}

		return handler(r)
	}
}

func NewGraphQLHandler(s *proto.Schema, api *proto.Api) Handler {
	gqlSchema, err := NewGraphQLSchema(s, api)
	if err != nil {
		panic(err)
	}

	return func(r *Request) (*Response, error) {
		var params struct {
			Query         string                 `json:"query"`
			OperationName string                 `json:"operationName"`
			Variables     map[string]interface{} `json:"variables"`
		}

		err := json.Unmarshal(r.Body, &params)
		if err != nil {
			return nil, err
		}

		result := graphql.Do(graphql.Params{
			Schema:         *gqlSchema,
			Context:        r.Context,
			RequestString:  params.Query,
			VariableValues: params.Variables,
		})

		b, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}

		return &Response{
			Body:   b,
			Status: 200,
		}, nil
	}
}
