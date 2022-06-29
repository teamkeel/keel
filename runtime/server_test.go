package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
)

func TestJustToDriveInitialCodingAndCompiling(t *testing.T) {

	schemaDir := filepath.Join(".", "testdata")
	s2m := schema.Builder{}
	protoSchema, err := s2m.MakeFromDirectory(schemaDir)
	require.NoError(t, err)
	protoJSON, err := json.Marshal(protoSchema)
	require.NoError(t, err)

	svr, err := NewServer(string(protoJSON))
	require.NoError(t, err)

	defer svr.Shutdown(context.Background())
	go svr.ListenAndServe()

	posturl := "http://localhost:8080/graphql/Web"
	req, err := http.NewRequest("POST", posturl, bytes.NewBuffer([]byte(exampleQuery)))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	httpResponse, err := client.Do(req)
	require.NoError(t, err)
	defer httpResponse.Body.Close()
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	responseBody, err := ioutil.ReadAll(httpResponse.Body)
	require.Nil(t, err)
	bodyStr := string(responseBody)

	const expected = "{\"data\":{\"getAuthor\":{\"name\":\"Harriet\"}}}"
	require.Equal(t, expected, bodyStr)
}

var exampleQuery string = `
{
	"query": "{ getAuthor(name: \"fred\") { name } }"
}
`
