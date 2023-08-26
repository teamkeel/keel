package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	types "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/teamkeel/keel/functions"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func NewFunctionsTransport(functionName string) functions.Transport {
	return func(ctx context.Context, request *functions.FunctionsRuntimeRequest) (*functions.FunctionsRuntimeResponse, error) {
		return invokeFunction(ctx, functionName, request)
	}
}

type LambdaErrorResponse struct {
	ErrorType    string   `json:"error_type,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
	Trace        []string `json:"trace,omitempty"`
}

func invokeFunction(ctx context.Context, functionName string, request *functions.FunctionsRuntimeRequest) (*functions.FunctionsRuntimeResponse, error) {
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
		FunctionName:   &functionName,
		InvocationType: types.InvocationTypeRequestResponse,
		LogType:        types.LogTypeNone,
		Payload:        requestJson,
	})
	if err != nil {
		fmt.Println("error from invoke", err.Error())
		return nil, err
	}

	fmt.Println("function response:", invokeOutput.StatusCode, string(invokeOutput.Payload))

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
			Result: lambdaResponse.Trace,
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
