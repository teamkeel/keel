package runtime

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/infra"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/sst-go"
)

var (
	runtimeHandler http.Handler
	privateKey     *rsa.PrivateKey
)

func Start() {
	privateKeyPem, err := sst.Secret(context.Background(), "KEEL_PRIVATE_KEY")
	if err != nil {
		panic(err)
	}

	privateKeyBlock, _ := pem.Decode([]byte(privateKeyPem))
	if privateKeyBlock == nil {
		panic(errors.New("private key pem cannot be decoded"))
	}

	privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		panic(err)
	}

	lambda.Start(handler)
}

func handler(ctx context.Context, event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	if runtimeHandler == nil {
		schema, err := infra.GetSchema(ctx)
		if err != nil {
			return events.LambdaFunctionURLResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       fmt.Sprintf("error loading schema: %s", err.Error()),
			}, nil
		}

		runtimeHandler = runtime.NewHttpHandler(schema)
	}

	database, err := infra.GetDatabase(ctx)
	if err != nil {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("error connecting to database: %s", err.Error()),
		}, nil
	}

	// TODO: add way of getting all secrets from sst-go
	ctx = runtimectx.WithSecrets(ctx, map[string]string{})
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	ctx = db.WithDatabase(ctx, database)

	functionsHandler, _ := sst.Function(ctx, "FunctionsHandler")
	if functionsHandler != nil {
		ctx = functions.WithFunctionsTransport(
			ctx,
			infra.NewFunctionsTransport(functionsHandler.FunctionName),
		)
	}

	runtimeRequest, err := eventToRequest(ctx, event)
	if err != nil {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("error converting lamdba event to request: %s", err.Error()),
		}, nil
	}

	w := core.NewProxyResponseWriterV2()

	runtimeHandler.ServeHTTP(http.ResponseWriter(w), runtimeRequest)

	resp, err := w.GetProxyResponse()
	if err != nil {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("error converting response: %s", err.Error()),
		}, nil
	}

	return events.LambdaFunctionURLResponse{
		StatusCode:      resp.StatusCode,
		Body:            resp.Body,
		Headers:         resp.Headers,
		Cookies:         resp.Cookies,
		IsBase64Encoded: resp.IsBase64Encoded,
	}, nil
}

// EventToRequest converts a Function URL event into an http.Request object.
// Returns the populated request maintaining headers
func eventToRequest(ctx context.Context, req events.LambdaFunctionURLRequest) (*http.Request, error) {
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, err
		}
		decodedBody = base64Body
	}

	path := req.RawPath
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	serverAddress := "https://" + req.RequestContext.DomainName
	path = serverAddress + path + "?" + req.RawQueryString

	httpRequest, err := http.NewRequest(
		strings.ToUpper(req.RequestContext.HTTP.Method),
		path,
		bytes.NewReader(decodedBody),
	)

	if err != nil {
		return nil, err
	}

	for header, val := range req.Headers {
		httpRequest.Header.Add(header, val)
	}

	httpRequest.RemoteAddr = req.RequestContext.HTTP.SourceIP
	httpRequest.RequestURI = httpRequest.URL.RequestURI()
	httpRequest = httpRequest.WithContext(ctx)

	return httpRequest, nil
}
