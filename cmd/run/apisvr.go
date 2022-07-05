package run

import (
	"fmt"

	"github.com/teamkeel/keel/runtime"
)

// startAPIServerAsync first creates an HTTP server that serves the GraphQL API(s), implied
// by the given (JSON serialized) proto.Schema. Then it launches a go routine in which
// to tell it to ListenAndServe.
func startAPIServerAsync(schemaJSON string) error {
	svr, err := runtime.NewServer(schemaJSON)
	if err != nil {
		return err
	}

	fmt.Printf("Your GraphQL server has been restarted. Serving your APIs at localhost:8080/graphql/<api-name>\n")

	// todo put in note about goroutine housekeeping
	go func() {
		err := svr.ListenAndServe()
		if err != nil {
			fmt.Printf("server ListenAndServe terminated with this error: %v", err)
		}
	}()

	return nil
}
