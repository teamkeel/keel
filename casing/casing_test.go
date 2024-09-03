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

func TestToSentenceCase(t *testing.T) {
	t.Parallel()
	testCases := []testCase{
		{
			input:  "slackID",
			output: "Slack id",
		},
		{
			input:  "slackId",
			output: "Slack id",
		},
		{
			input:  "SlackIDID",
			output: "Slack idid",
		},
		{
			input:  "APIKey",
			output: "Api key",
		},
		{
			input:  "snaked_AND_SCREAMING_case",
			output: "Snaked and screaming case",
		},
		{
			input:  "kebab-case",
			output: "Kebab case",
		},
	}

	for _, testCase := range testCases {
		actual := casing.ToSentenceCase(testCase.input)

		assert.Equal(t, testCase.output, actual)
	}
}

func TestToPlural(t *testing.T) {
	t.Parallel()
	testCases := []testCase{
		{input: "slack", output: "slacks"},
		{input: "person", output: "people"},
		{input: "wife", output: "wives"},
		{input: "sheep", output: "sheep"},
		{input: "sheep", output: "sheep"},
		{input: "fox", output: "foxes"},
		{input: "cherries", output: "cherries"},
		{input: "cherry", output: "cherries"},
		{input: "mouse", output: "mice"},
		{input: "order", output: "orders"},
	}

	for _, testCase := range testCases {
		actual := casing.ToPlural(testCase.input)

		assert.Equal(t, testCase.output, actual)
	}
}
