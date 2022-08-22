package functions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HttpFunctionsClient struct {
	Port string
	Host string
}

func (h *HttpFunctionsClient) Request(ctx context.Context, actionName string, body map[string]any) (map[string]any, error) {
	b, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s/%s", h.Host, h.Port, actionName), bytes.NewReader(b))

	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	b, err = ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	response := map[string]any{}

	json.Unmarshal(b, &response)

	return response, nil
}
