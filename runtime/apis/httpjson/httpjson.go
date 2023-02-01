package httpjson

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
)

type HttpJsonErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewHandler(p *proto.Schema, api *proto.Api) common.ApiHandlerFunc {
	return func(r *http.Request) common.Response {

		pathParts := strings.Split(r.URL.Path, "/")
		actionName := pathParts[len(pathParts)-1]
		var inputs map[string]any

		switch r.Method {
		case http.MethodGet:
			inputs = parseQueryParams(r.URL.Query())
		case http.MethodPost:
			var err error
			inputs, err = parsePostBody(r.Body)
			if err != nil {
				return common.NewJsonResponse(http.StatusInternalServerError, HttpJsonErrorResponse{
					Code:    "ERR_INTERNAL",
					Message: "error parsing POST body",
				})
			}
		default:
			return common.NewJsonResponse(http.StatusMethodNotAllowed, HttpJsonErrorResponse{
				Code:    "ERR_HTTP_METHOD_NOT_ALLOWED",
				Message: "only HTTP POST or GET accepted",
			})
		}

		op := proto.FindOperation(p, actionName)
		if op == nil {
			return common.NewJsonResponse(http.StatusNotFound, HttpJsonErrorResponse{
				Code:    "ERR_NOT_FOUND",
				Message: "method not found",
			})
		}

		scope := actions.NewScope(r.Context(), op, p)

		response, err := actions.Execute(scope, inputs)
		if err != nil {
			code := "ERR_INTERNAL"
			message := "error executing request"
			httpCode := http.StatusInternalServerError

			var runtimeErr common.RuntimeError
			if errors.As(err, &runtimeErr) {
				code = runtimeErr.Code
				message = runtimeErr.Message

				switch code {
				case common.ErrInvalidInput:
					httpCode = http.StatusBadRequest
				case common.ErrRecordNotFound:
					httpCode = http.StatusNotFound
				case common.ErrPermissionDenied:
					httpCode = http.StatusUnauthorized
				}
			}

			return common.NewJsonResponse(httpCode, HttpJsonErrorResponse{
				Code:    code,
				Message: message,
			})
		}

		return common.NewJsonResponse(http.StatusOK, response)
	}
}

func parseQueryParams(q url.Values) map[string]any {
	inputs := map[string]any{}
	for k := range q {
		inputs[k] = q.Get(k)
	}
	return inputs
}

func parsePostBody(b io.ReadCloser) (inputs map[string]any, err error) {
	body, err := io.ReadAll(b)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &inputs)
	return inputs, err
}
