package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
	"gorm.io/gorm"
)

func TestJustToDriveInitialCodingAndCompiling(t *testing.T) {

	schemaDir := filepath.Join(".", "testdata", "get-simplest-happy")
	s2m := schema.Builder{}
	protoSchema, err := s2m.MakeFromDirectory(schemaDir)
	require.NoError(t, err)
	protoJSON, err := json.Marshal(protoSchema)
	require.NoError(t, err)

	// todo set up the database properly
	var gormDB *gorm.DB = nil

	svr, err := NewServer(string(protoJSON), gormDB)
	require.NoError(t, err)

	defer svr.Shutdown(context.Background())
	go svr.ListenAndServe()

	// Avoid intermittend test failures - noted on CI
	time.Sleep(500 * time.Millisecond)

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

	const expected = "{\"data\":{\"getAuthor\":{\"name\":\"my-string\"}}}"
	require.Equal(t, expected, bodyStr)
}

var exampleQuery string = `
{
	"query": "{ getAuthor(name: \"fred\") { name } }"
}
`
