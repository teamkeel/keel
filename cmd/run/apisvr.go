package run

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/runtime"
)

// restartAPIServer stops and then restarts the GraphQL API server referenced by
// the handler's apiServer field. It is safe to call it when the server has not
// been started yet.
func (h *SchemaChangedHandler) retartAPIServer(schemaJSON string) (err error) {
	if h.apiServer != nil {
		err := h.apiServer.Shutdown(context.Background())
		if err != nil {
			return err
		}
	}

	h.apiServer, err = runtime.NewServer(schemaJSON)
	if err != nil {
		return err
	}

	// todo put in note about goroutine housekeeping
	go func() {
		err := h.apiServer.ListenAndServe()
		if err != nil {
			fmt.Printf("server ListenAndServe terminated with this error: %v", err)
		}
	}()

	fmt.Printf("Your GraphQL server has been restarted - serving your APIs at localhost:8080/graphql/<api-name>\n")
	return nil
}
