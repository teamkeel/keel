package functions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlPath "net/url"

	"github.com/pkg/errors"
)

type HttpFunctionsClient struct {
	URL string
}

func NewHttpTransport(url string) Transport {
	return func(ctx context.Context, req *FunctionsRuntimeRequest, functionType FunctionType) (*FunctionsRuntimeResponse, error) {
		b, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		targetUrl, err := urlPath.JoinPath(url, string(functionType))
		if err != nil {
			return nil, err
		}

		resp, err := http.Post(targetUrl, "application/json", bytes.NewReader(b))
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("node server responded with status code '%d' for '%s'", resp.StatusCode, targetUrl)
		}

		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var response FunctionsRuntimeResponse

		err = json.Unmarshal(b, &response)
		if err != nil {
			return nil, errors.New("invalid json: " + string(b))
		}

		return &response, nil
	}
}
