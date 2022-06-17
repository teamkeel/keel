package str

import (
	"github.com/gertd/go-pluralize"
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

var pl = pluralize.NewClient()

func IsSingular(str string) bool {
	return pl.IsSingular(str)
}

func IsPlural(str string) bool {
	return pl.IsPlural(str)
}

func Pluralize(str string) string {
	return pl.Pluralize(str, 1, false)
}

func Singularize(str string) string {
	return pl.Singular(str)
}
