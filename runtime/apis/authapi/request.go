package authapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func parsePostData(r *http.Request) (map[string]string, error) {
	data := map[string]string{}
	switch {
	case HasContentType(r.Header, "application/x-www-form-urlencoded"):
		_ = r.ParseForm()
		for k, v := range r.Form {
			data[k] = v[0]
		}
	case HasContentType(r.Header, "application/json"):
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		if string(body) == "" {
			break
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("content-type not provided or has no implementation")
	}

	return data, nil
}
