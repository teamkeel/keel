package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/teamkeel/keel/proto"
)

type FunctionsClient interface {
	Call(ctx context.Context, actionName string, opType proto.OperationType, body map[string]any) (any, error)
	ToGraphQL(ctx context.Context, response any, opType proto.OperationType) (interface{}, error)
}

type HttpFunctionsClient struct {
	Port string
	Host string
}

func (c *HttpFunctionsClient) Call(ctx context.Context, actionName string, opType proto.OperationType, body map[string]any) (any, error) {
	b, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s/%s", c.Host, c.Port, actionName), bytes.NewReader(b))

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

	var response interface{}

	json.Unmarshal(b, &response)

	return response, nil
}
