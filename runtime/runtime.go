package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/graphql-go/graphql"
	log "github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/runtime/actions"
	gql "github.com/teamkeel/keel/runtime/apis/graphql"
	rpc "github.com/teamkeel/keel/runtime/apis/rpc"
	"github.com/teamkeel/keel/runtime/common"

	"github.com/gorilla/handlers"
	"github.com/rs/cors"
	"github.com/teamkeel/keel/proto"

	"github.com/teamkeel/keel/runtime/runtimectx"
)

const (
	authorizationHeaderName string = "Authorization"
)

type Handler func(r *http.Request) (*common.Response, error)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(logLevel())
}

func Serve(currSchema *proto.Schema) func(w http.ResponseWriter, r *http.Request) {

	h := func(w http.ResponseWriter, r *http.Request) {

		log.WithFields(log.Fields{
			"url":     r.URL,
			"uri":     r.RequestURI,
			"headers": r.Header,
			"method":  r.Method,
			"host":    r.Host,
		}).Debug("request received")

		if currSchema == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Cannot serve requests when schema contains errors"))
			return
		}

		handler := NewHandler(currSchema)

		ctx := r.Context()

		header := r.Header.Get(authorizationHeaderName)
		if header != "" {
			headerSplit := strings.Split(header, "Bearer ")
			if len(headerSplit) != 2 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("no 'Bearer' prefix in the authentication header"))
				return
			}

			identityId, err := actions.ParseBearerToken(headerSplit[1])

			switch {
			case errors.Is(err, actions.ErrInvalidToken) || errors.Is(err, actions.ErrTokenExpired):
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error()))
				return
			case errors.Is(err, actions.ErrInvalidIdentityClaim):
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			// Check that identity actually does exist as it could
			// have been deleted after the bearer token was generated.
			identity, err := actions.FindIdentityById(ctx, identityId)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			if identity == nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(actions.ErrIdentityNotFound.Error()))
				return
			}

			ctx = runtimectx.WithIdentity(ctx, identityId)
			r = r.WithContext(ctx)
		}

		response, err := handler(r)

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
	withCors := cors.New(cors.Options{
		AllowOriginFunc: CheckOrigin,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(handler)

	return handlers.CompressHandler(withCors).ServeHTTP
}

func CheckOrigin(origin string) bool {
	// Returning true reflects the request origin as the allows origin to allow support of any origin along with AllowCredentials
	return true
}

func NewHandler(s *proto.Schema) Handler {
	handlers := map[string]Handler{}

	for _, api := range s.Apis {
		apiPath := strings.ToLower(api.Name)
		switch api.Type {
		case proto.ApiType_API_TYPE_GRAPHQL:
			handlers[apiPath] = NewGraphQLHandler(s, api)
		case proto.ApiType_API_TYPE_RPC:
			handlers[apiPath] = NewRpcHandler(s, api)
		default:
			panic(fmt.Sprintf("api type %s not supported", api.Type.String()))
		}
	}

	return func(r *http.Request) (*common.Response, error) {
		uriSegments := strings.Split(r.URL.Path, "/")
		apiPath := strings.ToLower(uriSegments[1])

		handler, ok := handlers[apiPath]
		if !ok {
			return &common.Response{
				Status: 404,
				Body:   []byte("Not found"),
			}, nil
		}

		return handler(r)
	}
}

func NewRpcHandler(s *proto.Schema, api *proto.Api) Handler {
	handler := rpc.NewRpcApi(s, api)

	return func(r *http.Request) (*common.Response, error) {
		status := 200

		data, err := handler(r)
		if err != nil {
			// TODO: add error codes
			data = map[string]any{
				"errors": []map[string]string{
					{
						"message": err.Error(),
					},
				},
			}
			status = 400
		}

		res, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		return &common.Response{
			Body:   res,
			Status: status,
		}, nil
	}
}

func NewGraphQLHandler(s *proto.Schema, api *proto.Api) Handler {
	gqlSchema, err := gql.NewGraphQLSchema(s, api)
	if err != nil {
		panic(err)
	}

	// This enables the graphql-go extension for tracing
	if os.Getenv("ENABLE_TRACING") == "true" {
		gqlSchema.AddExtensions(&Tracer{})
	}

	return func(r *http.Request) (*common.Response, error) {

		var params struct {
			Query         string                 `json:"query"`
			OperationName string                 `json:"operationName"`
			Variables     map[string]interface{} `json:"variables"`
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(body, &params)
		if err != nil {
			return nil, err
		}

		log.WithFields(log.Fields{
			"query": params.Query,
		}).Debug("graphql")

		result := graphql.Do(graphql.Params{
			Schema:         *gqlSchema,
			Context:        r.Context(),
			RequestString:  params.Query,
			VariableValues: params.Variables,
		})

		b, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}

		return &common.Response{
			Body:   b,
			Status: 200,
		}, nil
	}
}

func logLevel() log.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "trace":
		return log.TraceLevel
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}
