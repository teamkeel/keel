package formatting_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/formatting"
)

func TestHumanizeListDelimiterAnd(t *testing.T) {
	list := []string{"bananas", "apples", "oranges"}

	actual := formatting.HumanizeList(list, formatting.DelimiterAnd)

	assert.Equal(t, "bananas, apples, and oranges", actual)
}

func TestHumanizeListDelimiterOr(t *testing.T) {
	list := []string{"bananas", "apples", "oranges"}

	actual := formatting.HumanizeList(list, formatting.DelimiterOr)

	assert.Equal(t, "bananas, apples, or oranges", actual)
}

func TestHumanizeListOneItem(t *testing.T) {
	list := []string{"bananas"}

	actual := formatting.HumanizeList(list, formatting.DelimiterOr)

	assert.Equal(t, "bananas", actual)
}

func TestHumanizeListTwoItems(t *testing.T) {
	list := []string{"bananas", "oranges"}

	actual := formatting.HumanizeList(list, formatting.DelimiterAnd)

	assert.Equal(t, "bananas, and oranges", actual)
}
