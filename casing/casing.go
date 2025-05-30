package casing

import (
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/samber/lo"
)

func ToLowerCamel(str string) string {
	return toCamelCase(str, true)
}

func ToCamel(str string) string {
	return toCamelCase(str, false)
}

func ToSnake(str string) string {
	return strcase.ToSnake(str)
}

func ToScreamingSnake(str string) string {
	return strcase.ToScreamingSnake(str)
}

func ToKebab(str string) string {
	return strcase.ToKebab(str)
}

// ToSentenceCase will return a given string as "Sentence case"; splitting all components into words and capitalising
// just the first char. e.g. `listOrder` becomes "List order".
func ToSentenceCase(str string) string {
	s := strings.ReplaceAll(strcase.ToSnake(str), "_", " ")
	return capitalizeWord(s)
}

// To Plural will return the plural word for the given singular word.
func ToPlural(word string) string {
	return inflection.Plural(word)
}

func toCamelCase(input string, lowerCamel bool) string {
	words := camelcase.Split(input)

	// filter out any non letter chars such as '_' from the word array
	words = lo.Filter(words, func(word string, _ int) bool {
		runes := []rune(word)

		return lo.EveryBy(runes, func(r rune) bool {
			// the parser only allows for letter and number identifiers for models and fields
			return unicode.IsLetter(r) || unicode.IsNumber(r)
		})
	})

	str := ""

	for i, word := range words {
		if i == 0 && lowerCamel {
			str += strings.ToLower(word)

			continue
		}

		str += capitalizeWord(word)
	}

	return str
}

func capitalizeWord(word string) string {
	if len(word) < 1 {
		return word
	}
	return strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
}
