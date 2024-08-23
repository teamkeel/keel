package casing_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/casing"
)

type testCase struct {
	input, output string
}

func TestLowerCamel(t *testing.T) {
	t.Parallel()

	testCases := []testCase{
		{
			input:  "slackID",
			output: "slackId",
		},
		{
			input:  "slackId",
			output: "slackId",
		},
		{
			input:  "slackIDID",
			output: "slackIdid",
		},
		{
			input:  "APIKey",
			output: "apiKey",
		},
		{
			input:  "slack_API_Key",
			output: "slackApiKey",
		},
		{
			input:  "smallBIG",
			output: "smallBig",
		},
	}

	for _, testCase := range testCases {
		actual := casing.ToLowerCamel(testCase.input)

		assert.Equal(t, testCase.output, actual)
	}
}

func TestCamel(t *testing.T) {
	t.Parallel()
	testCases := []testCase{
		{
			input:  "slackID",
			output: "SlackId",
		},
		{
			input:  "slackId",
			output: "SlackId",
		},
		{
			input:  "slackIDID",
			output: "SlackIdid",
		},
		{
			input:  "APIKey",
			output: "ApiKey",
		},
	}

	for _, testCase := range testCases {
		actual := casing.ToCamel(testCase.input)

		assert.Equal(t, testCase.output, actual)
	}
}
