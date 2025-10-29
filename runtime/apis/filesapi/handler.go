package filesapi

import (
	"bytes"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/apis/tasksapi")

// Handler returns a HTTP handler used to return files fromt he database storage:
// GET files/{id} - Retrieves the file with the given ID.
func NewFilesHandler(schema *proto.Schema) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "FilesAPI")
		defer span.End()

		path := path.Clean(r.URL.EscapedPath())
		pathParts := strings.Split(strings.TrimPrefix(path, "/files/"), "/")

		if len(pathParts) != 1 {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}

		store, err := runtimectx.GetStorage(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, fi, err := store.GetFileData(pathParts[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if data == nil {
			http.NotFound(w, r)
			return
		}
		reader := bytes.NewReader(data)
		w.Header().Set("Content-Type", fi.ContentType)
		w.Header().Set("Content-Disposition", `attachment; filename="`+fi.Filename+`"`)

		http.ServeContent(w, r, fi.Filename, time.Now(), reader)
	}
}
