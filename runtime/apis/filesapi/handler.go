package filesapi

import (
	"net/http"
	"path"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/apis/tasksapi")

// Handler returns a HTTP handler used to return files fromt he database storage:
// GET files/{id} - Retrieves the file with the given ID.
func NewFilesHandler(schema *proto.Schema) func(http.ResponseWriter, *http.Request) common.Response {
	return func(w http.ResponseWriter, r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "FilesAPI")
		defer span.End()

		path := path.Clean(r.URL.EscapedPath())
		pathParts := strings.Split(strings.TrimPrefix(path, "/files/"), "/")

		if len(pathParts) != 1 {
			return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
		}

		if r.Method != http.MethodGet {
			return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET accepted"), nil)
		}

		store, err := runtimectx.GetStorage(ctx)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		data, fi, err := store.GetFileData(pathParts[0])
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
		}

		if data == nil {
			return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
		}

		return common.Response{
			Status: http.StatusOK,
			Body:   data,
			Headers: map[string][]string{
				"Content-Type":        {fi.ContentType},
				"Content-Disposition": {"attachment; filename=" + fi.Filename},
			},
		}
	}
}
