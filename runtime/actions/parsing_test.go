package actions_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/types"
)

func TestParsing(t *testing.T) {
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

	parsed, err := actions.TransformInputTypes(scope.Schema, action, data)
	assert.NoError(t, err)

	assert.IsType(t, "", parsed["title"])
	assert.IsType(t, 0, parsed["views"])
	assert.IsType(t, types.Date{}, parsed["postAt"])
	assert.IsType(t, types.Timestamp{}, parsed["created"])
}

func TestParsingArrays(t *testing.T) {
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

	parsed, err := actions.TransformInputTypes(scope.Schema, action, data)
	assert.NoError(t, err)

	assert.IsType(t, []string{}, parsed["texts"])
	assert.IsType(t, []int{}, parsed["numbers"])
	assert.IsType(t, []types.Date{}, parsed["dates"])
	assert.IsType(t, []types.Timestamp{}, parsed["timestamps"])
}

func TestParsingUpdate(t *testing.T) {
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

	parsed, err := actions.TransformInputTypes(scope.Schema, action, data)
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

	parsed, err := actions.TransformInputTypes(scope.Schema, action, data)
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

	parsed, err := actions.TransformInputTypes(scope.Schema, action, data)
	assert.NoError(t, err)

	assert.Equal(t, "Keel post", parsed["title"])
	assert.IsType(t, map[string]any{}, parsed["author"])

	author := parsed["author"].(map[string]any)
	assert.Equal(t, "Keelson", author["name"])
}
