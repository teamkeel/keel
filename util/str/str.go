package str

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func PadRight(str string, padAmount int) string {
	for len(str) < padAmount {
		str += " "
	}

	return str
}

func AsTitle(str string) string {
	caser := cases.Title(language.English)
	return caser.String(str)
}
