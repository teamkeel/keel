package main

import (
	"fmt"

	"context"
	"encoding/json"

	"github.com/teamkeel/keel/functions"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func NewLambdaInvokeTransport(functionsArn string) functions.Transport {
	return func(ctx context.Context, request *functions.FunctionsRuntimeRequest) (*functions.FunctionsRuntimeResponse, error) {
		return invokeFunction(ctx, functionsArn, request)
	}
}

type LambdaErrorResponse struct {
	ErrorType    string   `json:"errorType,omitempty"`
	ErrorMessage string   `json:"errorMessage,omitempty"`
	Trace        []string `json:"stackTrace,omitempty"`
}

func invokeFunction(ctx context.Context, functionArn string, request *functions.FunctionsRuntimeRequest) (*functions.FunctionsRuntimeResponse, error) {
	span := trace.SpanFromContext(ctx)

	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	invokeOutput, err := lambda.NewFromConfig(cfg).Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   &functionArn,
		InvocationType: lambdaTypes.InvocationTypeRequestResponse,
		LogType:        lambdaTypes.LogTypeNone,
		Payload:        requestJson,
	})
	if err != nil {
		return nil, err
	}

	if invokeOutput.StatusCode < 200 || invokeOutput.StatusCode >= 300 {
		err = fmt.Errorf("non-200 (%d) status from function", invokeOutput.StatusCode)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Int("function.response_status", int(invokeOutput.StatusCode)))

		if invokeOutput.FunctionError != nil {
			span.SetAttributes(attribute.String("function.response_error", *invokeOutput.FunctionError))
		}

		if len(invokeOutput.Payload) > 0 {
			span.SetAttributes(attribute.String("function.response_payload", string(invokeOutput.Payload)))
		}

		return nil, err
	}

	response := functions.FunctionsRuntimeResponse{
		Result: nil,
		Error:  nil,
	}

	if invokeOutput.FunctionError != nil {
		var lambdaResponse LambdaErrorResponse

		err = json.Unmarshal(invokeOutput.Payload, &lambdaResponse)
		if err != nil {
			return nil, err
		}

		response = functions.FunctionsRuntimeResponse{
			Error: &functions.FunctionsRuntimeError{
				Message: lambdaResponse.ErrorMessage,
				Code:    functions.UnknownError,
			},
		}
	}

	if invokeOutput.Payload != nil {
		err = json.Unmarshal(invokeOutput.Payload, &response)
		if err != nil {
			return nil, err
		}
	}

	return &response, nil
}
