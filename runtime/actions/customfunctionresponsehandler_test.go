package actions_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

type TestFunctionsClient struct {
	testData any
}

func (c *TestFunctionsClient) Request(ctx context.Context, actionName string, opType proto.OperationType, body map[string]any) (any, error) {
	d := c.testData

	c.testData = nil

	return d, nil
}

func (c *TestFunctionsClient) ToGraphQL(ctx context.Context, response any, opType proto.OperationType) (interface{}, error) {
	return nil, nil
}

func TestCustomFunctionGetResponseTransformation(t *testing.T) {
	response := map[string]any{
		"object": map[string]any{
			"id":    "123",
			"aDate": "2022-11-29T15:47:22.951Z",
		},
	}

	client := &TestFunctionsClient{
		testData: response,
	}

	ctx := context.WithValue(context.Background(), functions.FunctionsClientKey, client)

	op := &proto.Operation{}

	result, err := actions.ParseGetObjectResponse(ctx, op, map[string]any{})

	assert.NoError(t, err)

	assert.IsType(t, time.Now(), result["aDate"])

	assert.Equal(t, result["aDate"], time.Date(2022, 11, 29, 15, 47, 22, 951000000, time.UTC))
}

func TestCustomFunctionCreateResponseTransformation(t *testing.T) {
	response := map[string]any{
		"object": map[string]any{
			"id":    "123",
			"aDate": "2022-11-29T15:47:22.951Z",
		},
	}

	client := &TestFunctionsClient{
		testData: response,
	}

	ctx := context.WithValue(context.Background(), functions.FunctionsClientKey, client)

	op := &proto.Operation{}

	result, err := actions.ParseCreateObjectResponse(ctx, op, map[string]any{})

	assert.NoError(t, err)

	assert.IsType(t, time.Now(), result["aDate"])

	assert.Equal(t, result["aDate"], time.Date(2022, 11, 29, 15, 47, 22, 951000000, time.UTC))
}

func TestCustomFunctionUpdateResponseTransformation(t *testing.T) {
	response := map[string]any{
		"object": map[string]any{
			"id":    "123",
			"aDate": "2022-11-29T15:47:22.951Z",
		},
	}

	client := &TestFunctionsClient{
		testData: response,
	}

	ctx := context.WithValue(context.Background(), functions.FunctionsClientKey, client)

	op := &proto.Operation{}

	result, err := actions.ParseUpdateResponse(ctx, op, map[string]any{})

	assert.NoError(t, err)

	assert.IsType(t, time.Now(), result["aDate"])

	assert.Equal(t, result["aDate"], time.Date(2022, 11, 29, 15, 47, 22, 951000000, time.UTC))
}
