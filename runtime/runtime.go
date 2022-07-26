package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"gorm.io/gorm"
)

type Request struct {
	Context context.Context
	URL     url.URL
	Body    []byte
}

type Response struct {
	Body   []byte
	Status int
}

type Handler func(r *Request) (*Response, error)

func NewHandler(db *gorm.DB, s *proto.Schema) Handler {
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
		handler, ok := handlers[r.URL.Path]
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
