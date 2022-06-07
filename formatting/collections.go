package formatting

import "fmt"

type Delimiter string

const (
	DelimiterAnd Delimiter = "and"
	DelimiterOr  Delimiter = "or"
)

func HumanizeList(list []string, lastItemDelimiter Delimiter) string {
	strLength := len(list)
	output := ""

	for i, item := range list {
		if i < strLength-1 {
			output += fmt.Sprintf("%s, ", item)
		} else {
			output += fmt.Sprintf("%s %s", lastItemDelimiter, item)
		}
	}

	return output
}
