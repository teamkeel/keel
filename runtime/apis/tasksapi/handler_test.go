package tasksapi_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/tasks"
	keeltesting "github.com/teamkeel/keel/testing"
)

var taskTestSchema = `
task TestTask {
    fields {
        name Text
    }
    @permission(expression: ctx.isAuthenticated)
}
`

func TestUnassignTask_Success(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), taskTestSchema, true)
	defer database.Close()

	// Create an identity
	identity, err := actions.CreateIdentity(ctx, schema, "test@keel.xyz", "1234", oauth.KeelIssuer)
	require.NoError(t, err)
	identityID := identity["id"].(string)

	// Generate access token
	accessToken, _, err := oauth.GenerateAccessToken(ctx, identityID)
	require.NoError(t, err)

	// Get the task definition from schema
	pbTask := schema.FindTask("TestTask")
	require.NotNil(t, pbTask)

	// Create a task
	task, err := tasks.NewTask(ctx, pbTask, identityID, nil, map[string]any{"name": "Test Task"})
	require.NoError(t, err)

	// Assign the task
	task, err = tasks.AssignTask(ctx, pbTask, task.ID, identityID, identityID)
	require.NoError(t, err)
	require.Equal(t, tasks.StatusAssigned, task.Status)
	require.NotNil(t, task.AssignedTo)

	// Make unassign request
	request := makeUnassignRequest(ctx, "TestTask", task.ID, accessToken)
	response, httpResponse, err := handleRuntimeRequest[tasks.Task](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// Verify task is unassigned
	require.Equal(t, tasks.StatusNew, response.Status)
	require.Nil(t, response.AssignedTo)
	require.Nil(t, response.AssignedAt)
}

func TestUnassignTask_CannotUnassignCompletedTask(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), taskTestSchema, true)
	defer database.Close()

	// Create an identity
	identity, err := actions.CreateIdentity(ctx, schema, "test@keel.xyz", "1234", oauth.KeelIssuer)
	require.NoError(t, err)
	identityID := identity["id"].(string)

	// Generate access token
	accessToken, _, err := oauth.GenerateAccessToken(ctx, identityID)
	require.NoError(t, err)

	// Get the task definition from schema
	pbTask := schema.FindTask("TestTask")
	require.NotNil(t, pbTask)

	// Create a task
	task, err := tasks.NewTask(ctx, pbTask, identityID, nil, map[string]any{"name": "Test Task"})
	require.NoError(t, err)

	// Complete the task
	task, err = tasks.CompleteTask(ctx, pbTask, task.ID, identityID)
	require.NoError(t, err)
	require.Equal(t, tasks.StatusCompleted, task.Status)

	// Try to unassign the completed task - should fail
	request := makeUnassignRequest(ctx, "TestTask", task.ID, accessToken)
	errorResponse, httpResponse, err := handleRuntimeRequest[ErrorResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Contains(t, errorResponse.Message, "cannot unassign a completed or cancelled task")
}

func TestUnassignTask_CannotUnassignCancelledTask(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), taskTestSchema, true)
	defer database.Close()

	// Create an identity
	identity, err := actions.CreateIdentity(ctx, schema, "test@keel.xyz", "1234", oauth.KeelIssuer)
	require.NoError(t, err)
	identityID := identity["id"].(string)

	// Generate access token
	accessToken, _, err := oauth.GenerateAccessToken(ctx, identityID)
	require.NoError(t, err)

	// Get the task definition from schema
	pbTask := schema.FindTask("TestTask")
	require.NotNil(t, pbTask)

	// Create a task
	task, err := tasks.NewTask(ctx, pbTask, identityID, nil, map[string]any{"name": "Test Task"})
	require.NoError(t, err)

	// Cancel the task
	task, err = tasks.CancelTask(ctx, pbTask, task.ID, identityID)
	require.NoError(t, err)
	require.Equal(t, tasks.StatusCancelled, task.Status)

	// Try to unassign the cancelled task - should fail
	request := makeUnassignRequest(ctx, "TestTask", task.ID, accessToken)
	errorResponse, httpResponse, err := handleRuntimeRequest[ErrorResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Contains(t, errorResponse.Message, "cannot unassign a completed or cancelled task")
}

func TestUnassignTask_TaskNotFound(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), taskTestSchema, true)
	defer database.Close()

	// Create an identity
	identity, err := actions.CreateIdentity(ctx, schema, "test@keel.xyz", "1234", oauth.KeelIssuer)
	require.NoError(t, err)
	identityID := identity["id"].(string)

	// Generate access token
	accessToken, _, err := oauth.GenerateAccessToken(ctx, identityID)
	require.NoError(t, err)

	// Try to unassign a non-existent task
	request := makeUnassignRequest(ctx, "TestTask", "non-existent-id", accessToken)
	_, httpResponse, err := handleRuntimeRequest[ErrorResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, httpResponse.StatusCode)
}

func TestUnassignTask_UnassignNewTask(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), taskTestSchema, true)
	defer database.Close()

	// Create an identity
	identity, err := actions.CreateIdentity(ctx, schema, "test@keel.xyz", "1234", oauth.KeelIssuer)
	require.NoError(t, err)
	identityID := identity["id"].(string)

	// Generate access token
	accessToken, _, err := oauth.GenerateAccessToken(ctx, identityID)
	require.NoError(t, err)

	// Get the task definition from schema
	pbTask := schema.FindTask("TestTask")
	require.NotNil(t, pbTask)

	// Create a task (NEW status, not assigned)
	task, err := tasks.NewTask(ctx, pbTask, identityID, nil, map[string]any{"name": "Test Task"})
	require.NoError(t, err)
	require.Equal(t, tasks.StatusNew, task.Status)

	// Unassign a NEW task - should succeed (idempotent)
	request := makeUnassignRequest(ctx, "TestTask", task.ID, accessToken)
	response, httpResponse, err := handleRuntimeRequest[tasks.Task](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.Equal(t, tasks.StatusNew, response.Status)
}

func TestUnassignTask_HttpMethodNotAllowed(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), taskTestSchema, true)
	defer database.Close()

	// Create an identity
	identity, err := actions.CreateIdentity(ctx, schema, "test@keel.xyz", "1234", oauth.KeelIssuer)
	require.NoError(t, err)
	identityID := identity["id"].(string)

	// Generate access token
	accessToken, _, err := oauth.GenerateAccessToken(ctx, identityID)
	require.NoError(t, err)

	// Make GET request instead of PUT
	request := httptest.NewRequest(http.MethodGet, "http://mykeelapp.keel.so/topics/json/TestTask/tasks/some-id/unassign", nil)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+accessToken)
	request = request.WithContext(ctx)

	_, httpResponse, err := handleRuntimeRequest[ErrorResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusMethodNotAllowed, httpResponse.StatusCode)
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func makeUnassignRequest(ctx context.Context, topicName string, taskID string, accessToken string) *http.Request {
	request := httptest.NewRequest(http.MethodPut, "http://mykeelapp.keel.so/topics/json/"+topicName+"/tasks/"+taskID+"/unassign", nil)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+accessToken)
	request = request.WithContext(ctx)
	return request
}

func handleRuntimeRequest[T any](schema *proto.Schema, req *http.Request) (T, *http.Response, error) {
	var response T
	handler := runtime.NewHttpHandler(schema)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	httpResponse := w.Result()

	data, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return response, nil, err
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return response, httpResponse, nil
	}

	return response, httpResponse, nil
}
