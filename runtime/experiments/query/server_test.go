package main

import (
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

	resp, err := http.Get(`http://localhost:8080/graphql?query={user(id:"1"){name}}`)
	require.Nil(t, err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	bodyStr := string(body)

	const expected = "{\"data\":{\"user\":{\"name\":\"fred\"}}}\n"
	require.Equal(t, expected, bodyStr)
}
