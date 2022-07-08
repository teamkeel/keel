package run

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
)

// restartAPIServer stops and then restarts the GraphQL API server referenced by
// the handler's apiServer field. It is safe to call it when the server has not
// been started yet.
func (h *SchemaChangedHandler) retartAPIServer(schema *proto.Schema) (err error) {
	if h.apiServer != nil {
		err := h.apiServer.Shutdown(context.Background())
		if err != nil {
			return err
		}
	}

	h.apiServer, err = runtime.NewServer(schema, h.gormDB)
	if err != nil {
		return err
	}

	// todo put in note about goroutine housekeeping
	go func() {
		err := h.apiServer.ListenAndServe()
		// todo: this is documented as always returning a non-nil error.
		// but that includes our routine case, when we restart the server.
		// So do we need to discrimate between this error case and other error cases?
		_ = err
	}()

	fmt.Printf("Your GraphQL server has been restarted - serving your APIs at localhost:8080/graphql/<api-name>\n")
	return nil
}
