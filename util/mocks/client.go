package mocks

import "net/http"

var (
	DoFunc  func(req *http.Request) (*http.Response, error)
	GetFunc func(string) (*http.Response, error)
)

func init() {
	GetFunc = func(url string) (*http.Response, error) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		return DoFunc(req)
	}
}

type MockClient struct {
	DoFunc  func(req *http.Request) (*http.Response, error)
	GetFunc func(string) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return DoFunc(req)
}

func (m *MockClient) Get(url string) (*http.Response, error) {
	return GetFunc(url)
}
