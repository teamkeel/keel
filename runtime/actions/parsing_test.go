package actions_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/types"
)

func TestParsing(t *testing.T) {
	t.Parallel()
	schema := `
model Post {
	fields {
		title Text
		views Number
		postAt Date
		created Timestamp
	}
	actions {
		create createPost() with (title, views, postAt, created)
	}
}`

	input := ` 
{
	"title": "Keel",
	"views": 10,
	"postAt": "2024-04-08",
	"created": "2024-04-08T12:14:59Z"
}`

	scope, _, action, err := generateQueryScope(context.Background(), schema, "createPost")
	assert.NoError(t, err)

	var data map[string]any
	err = json.Unmarshal([]byte(input), &data)
	assert.NoError(t, err)

	message := scope.Schema.FindMessage(action.GetInputMessageName())
	isFunction := action.GetImplementation() == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM

	parsed, err := actions.TransformInputs(scope.Schema, message, data, isFunction)
	assert.NoError(t, err)

	assert.IsType(t, "", parsed["title"])
	assert.IsType(t, 0, parsed["views"])
	assert.IsType(t, types.Date{}, parsed["postAt"])
	assert.IsType(t, types.Timestamp{}, parsed["created"])
}

func TestParsingArrays(t *testing.T) {
	t.Parallel()
	schema := `
model Post {
	fields {
		texts Text[]
		numbers Number[]
		dates Date[]
		timestamps Timestamp[]
	}
	actions {
		create createPost() with (texts, numbers, dates, timestamps)
	}
}`

	input := ` 
{
	"texts": ["Keel", "Weave"],
	"numbers": [1,2,3],
	"dates": ["2024-04-08", "2024-02-15"],
	"timestamps": ["2024-04-08T12:14:59Z", "2024-02-01T08:00:00Z"]
}`

	scope, _, action, err := generateQueryScope(context.Background(), schema, "createPost")
	assert.NoError(t, err)

	var data map[string]any
	err = json.Unmarshal([]byte(input), &data)
	assert.NoError(t, err)

	message := scope.Schema.FindMessage(action.GetInputMessageName())
	isFunction := action.GetImplementation() == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM

	parsed, err := actions.TransformInputs(scope.Schema, message, data, isFunction)
	assert.NoError(t, err)

	assert.IsType(t, []string{}, parsed["texts"])
	assert.IsType(t, []int{}, parsed["numbers"])
	assert.IsType(t, []types.Date{}, parsed["dates"])
	assert.IsType(t, []types.Timestamp{}, parsed["timestamps"])
}

func TestParsingUpdate(t *testing.T) {
	t.Parallel()
	schema := `
model Post {
	fields {
		title Text
		views Number
		postAt Date
		created Timestamp
	}
	actions {
		update updatePost(id) with (title, views, postAt, created)
	}
}`

	input := ` 
{
	"where": {
		"id": "id-of-post"
	},
	"values": {
		"title": "Keel",
		"views": 10,
		"postAt": "2024-04-08",
		"created": "2024-04-08T12:14:59Z"
	}
}`

	scope, _, action, err := generateQueryScope(context.Background(), schema, "updatePost")
	assert.NoError(t, err)

	var data map[string]any
	err = json.Unmarshal([]byte(input), &data)
	assert.NoError(t, err)

	message := scope.Schema.FindMessage(action.GetInputMessageName())
	isFunction := action.GetImplementation() == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM

	parsed, err := actions.TransformInputs(scope.Schema, message, data, isFunction)
	assert.NoError(t, err)

	assert.IsType(t, map[string]any{}, parsed["where"])
	assert.IsType(t, map[string]any{}, parsed["values"])

	where := parsed["where"].(map[string]any)
	values := parsed["values"].(map[string]any)

	assert.IsType(t, "", where["id"])
	assert.IsType(t, "", values["title"])
	assert.IsType(t, 0, values["views"])
	assert.IsType(t, types.Date{}, values["postAt"])
	assert.IsType(t, types.Timestamp{}, values["created"])
}

func TestParsingNestedMany(t *testing.T) {
	t.Parallel()
	schema := `
model Author {
	fields {
		name Text
		posts Post[]
	}
	actions {
		create createAuthor() with (name, posts.title, posts.views, posts.postAt, posts.created)
	}
}
model Post {
	fields {
		title Text
		views Number
		postAt Date
		created Timestamp
		author Author
	}
}`

	input := ` 
{
	"name": "Keelson",
	"posts": [
		{
			"title": "Keel post 1",
			"views": 10,
			"postAt": "2024-04-08",
			"created": "2024-04-08T12:14:59Z"
		},
		{
			"title": "Keel post 2",
			"views": 2,
			"postAt": "2024-03-13",
			"created": "2024-02-01T08:00:00Z"
		}
	]

}`

	scope, _, action, err := generateQueryScope(context.Background(), schema, "createAuthor")
	assert.NoError(t, err)

	var data map[string]any
	err = json.Unmarshal([]byte(input), &data)
	assert.NoError(t, err)

	message := scope.Schema.FindMessage(action.GetInputMessageName())
	isFunction := action.GetImplementation() == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM

	parsed, err := actions.TransformInputs(scope.Schema, message, data, isFunction)
	assert.NoError(t, err)

	assert.IsType(t, "", parsed["name"])
	assert.IsType(t, []any{}, parsed["posts"])

	posts := parsed["posts"].([]any)

	post1 := posts[0].(map[string]any)
	assert.Equal(t, "Keel post 1", post1["title"])
	assert.Equal(t, 10, post1["views"])
	assert.Equal(t, types.Date{Time: time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)}, post1["postAt"])
	assert.Equal(t, types.Timestamp{Time: time.Date(2024, 4, 8, 12, 14, 59, 0, time.UTC)}, post1["created"])

	post2 := posts[1].(map[string]any)
	assert.Equal(t, "Keel post 2", post2["title"])
	assert.Equal(t, 2, post2["views"])
	assert.Equal(t, types.Date{Time: time.Date(2024, 3, 13, 0, 0, 0, 0, time.UTC)}, post2["postAt"])
	assert.Equal(t, types.Timestamp{Time: time.Date(2024, 2, 1, 8, 0, 0, 0, time.UTC)}, post2["created"])
}

func TestParsingNestedOne(t *testing.T) {
	t.Parallel()
	schema := `
model Author {
	fields {
		name Text
		posts Post[]
	}
	
}
model Post {
	fields {
		title Text
		author Author
	}
	actions {
		create createPost() with (title, author.name)
	}
}`

	input := ` 
{
	"title": "Keel post",
	"author": {
		"name": "Keelson"
	}

}`

	scope, _, action, err := generateQueryScope(context.Background(), schema, "createPost")
	assert.NoError(t, err)

	var data map[string]any
	err = json.Unmarshal([]byte(input), &data)
	assert.NoError(t, err)

	message := scope.Schema.FindMessage(action.GetInputMessageName())
	isFunction := action.GetImplementation() == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM

	parsed, err := actions.TransformInputs(scope.Schema, message, data, isFunction)
	assert.NoError(t, err)

	assert.Equal(t, "Keel post", parsed["title"])
	assert.IsType(t, map[string]any{}, parsed["author"])

	author := parsed["author"].(map[string]any)
	assert.Equal(t, "Keelson", author["name"])
}

func TestParsingCustomFunctionInputs(t *testing.T) {
	t.Parallel()
	dataUrl := `data:image/png;name=my-avatar.png;base64,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`
	schema := `
model Person {
	fields {
		avatar File
	}
	actions {
		write setAvatar(FileInput) returns (FileResponse)
	}
}
message FileInput {
    file File
}

message FileResponse {
    filename Text
}
`

	input := ` 
{
	"file": "` + dataUrl + `"
}`

	scope, _, action, err := generateQueryScope(context.Background(), schema, "setAvatar")
	assert.NoError(t, err)

	var data map[string]any
	err = json.Unmarshal([]byte(input), &data)
	assert.NoError(t, err)

	message := scope.Schema.FindMessage(action.GetInputMessageName())
	isFunction := action.GetImplementation() == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM

	parsed, err := actions.TransformInputs(scope.Schema, message, data, isFunction)
	assert.NoError(t, err)

	assert.IsType(t, map[string]any{}, parsed["file"])

	file := parsed["file"].(map[string]any)
	assert.Equal(t, "InlineFile", file["__typename"])
	assert.Equal(t, dataUrl, file["dataURL"])
}
