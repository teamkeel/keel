package runtime

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/apis/graphql"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/runtime/apis/jsonrpc"
	"github.com/teamkeel/keel/runtime/common"

	"github.com/teamkeel/keel/proto"

	"github.com/teamkeel/keel/runtime/runtimectx"
)

const (
	authorizationHeaderName string = "Authorization"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(logLevel())
}

func NewHttpHandler(currSchema *proto.Schema) http.Handler {
	httpHandler := func(w http.ResponseWriter, r *http.Request) {

		log.WithFields(log.Fields{
			"url":     r.URL,
			"uri":     r.RequestURI,
			"headers": r.Header,
			"method":  r.Method,
			"host":    r.Host,
		}).Debug("request received")

		if currSchema == nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Cannot serve requests when schema contains errors"))
			return
		}

		handler := NewHandler(currSchema)

		ctx := r.Context()

		header := r.Header.Get(authorizationHeaderName)
		if header != "" {
			headerSplit := strings.Split(header, "Bearer ")
			if len(headerSplit) != 2 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("no 'Bearer' prefix in the authentication header"))
				return
			}

			identityId, err := actions.ParseBearerToken(headerSplit[1])

			switch {
			case errors.Is(err, actions.ErrInvalidToken) || errors.Is(err, actions.ErrTokenExpired):
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(err.Error()))
				return
			case errors.Is(err, actions.ErrInvalidIdentityClaim):
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			// Check that identity actually does exist as it could
			// have been deleted after the bearer token was generated.
			identity, err := actions.FindIdentityById(ctx, currSchema, identityId)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			if identity == nil {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(actions.ErrIdentityNotFound.Error()))
				return
			}

			ctx = runtimectx.WithIdentity(ctx, identity)
		}

		// Collect request headers and add to runtime context
		// These are exposed in custom functions and in expressions
		headers := map[string][]string{}
		for k := range r.Header {
			headers[k] = r.Header.Values(k)
		}
		ctx = runtimectx.WithRequestHeaders(ctx, headers)
		r = r.WithContext(ctx)

		response := handler(r)

		// Add any custom headers to response, and join
		// into a single string where multi values exists
		for k, values := range response.Headers {
			for _, value := range values {
				w.Header().Add(k, value)
			}
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(response.Status)
		_, _ = w.Write(response.Body)
	}

	cors := cors.New(cors.Options{
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
	})

	return cors.Handler(http.HandlerFunc(httpHandler))
}

func CheckOrigin(origin string) bool {
	// Returning true reflects the request origin as the allows origin to allow support of any origin along with AllowCredentials
	return true
}

func NewHandler(s *proto.Schema) common.ApiHandlerFunc {
	handlers := map[string]common.ApiHandlerFunc{}

	for _, api := range s.Apis {
		root := "/" + strings.ToLower(api.Name)

		handlers[root+"/graphql"] = graphql.NewHandler(s, api)
		handlers[root+"/rpc"] = jsonrpc.NewHandler(s, api)

		httpJson := httpjson.NewHandler(s, api)
		for _, name := range proto.GetActionNamesForApi(s, api) {
			handlers[root+"/json/"+name] = httpJson
		}
	}

	return func(r *http.Request) common.Response {
		handler, ok := handlers[r.URL.Path]
		if !ok {
			return common.Response{
				Status: 404,
				Body:   []byte("Not found"),
			}
		}

		return handler(r)
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
