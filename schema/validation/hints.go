package validation

import (
	"fmt"
	"strings"

	levenshtein "github.com/ka-weihe/fast-levenshtein"
)

type Hint interface {
	ToString() string
}

type CorrectionHint struct {
	Hint
	Query   string
	Results []string
}

func NewCorrectionHint(referenceCollection []string, query string) *CorrectionHint {
	matches := make([]string, 0)
	attributeNames := make([]string, 0)

	for _, item := range referenceCollection {
		attributeNames = append(attributeNames, item)

		if levenshtein.Distance(query, item) < 2 {
			matches = append(matches, item)
		}
	}

	if len(matches) < 1 {
		matches = append(matches, attributeNames...)
	}

	return &CorrectionHint{Results: matches, Query: query}
}

func (hint *CorrectionHint) ToString() string {
	message := ""

	if len(hint.Results) == 1 {
		message = fmt.Sprintf("Did you mean %s?", hint.Results[0])
	} else if len(hint.Results) > 1 {
		message = fmt.Sprintf("Did you mean one of %s?", strings.Join(hint.Results, ", "))
	}

	return message
}

type NormalHint struct {
	Hint
	Message string
}

func NewHint(message string) *NormalHint {
	return &NormalHint{Message: message}
}

func (hint *NormalHint) ToString() string {
	return hint.Message
}
