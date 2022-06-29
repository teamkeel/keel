package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	s := NewServer()
	defer s.Shutdown(context.Background())
	go s.ListenAndServe()

	posturl := "http://localhost:8080/graphql"
	requestBody := []byte(`{ "query": "mutation { createTodo(text:\"My New todo\") { id text done } }" }'`)
	req, err := http.NewRequest("POST", posturl, bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	httpResponse, err := client.Do(req)
	require.NoError(t, err)
	defer httpResponse.Body.Close()

	responseBody, err := ioutil.ReadAll(httpResponse.Body)
	require.Nil(t, err)
	bodyStr := string(responseBody)

	const expected = "{\"data\":{\"createTodo\":{\"done\":false,\"id\":\"999\",\"text\":\"My New todo\"}}}\n"
	require.Equal(t, expected, bodyStr)
}
