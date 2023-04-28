package casing

import (
	"strings"
)

// Add more uppercase 'special cases' here
var uppercaseTransforms = map[string]string{
	"ID":  "Id",
	"API": "Api",
}

func ToLowerCamel(str string) string {
	return toCamelInitCase(str, false)
}

func ToCamel(str string) string {
	return toCamelInitCase(str, true)
}

func toCamelInitCase(s string, initCase bool) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// if the string is exactly equal to a value in uppercase transforms
	// then we want to just use the equivalent camel cased version
	if a, ok := uppercaseTransforms[s]; ok {
		s = a
	}

	// loop over all of the uppercase acronyms to be transformed
	for key, transform := range uppercaseTransforms {
		if strings.Contains(s, key) {
			s = strings.ReplaceAll(s, key, transform)
		}
	}

	n := strings.Builder{}
	n.Grow(len(s))
	capNext := initCase

	for i, v := range []byte(s) {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'
		if capNext {
			if vIsLow {
				v += 'A'
				v -= 'a'
			}
		} else if i == 0 {
			if vIsCap {
				v += 'a'
				v -= 'A'
			}
		}
		if vIsCap || vIsLow {
			n.WriteByte(v)
			capNext = false
		} else if vIsNum := v >= '0' && v <= '9'; vIsNum {
			n.WriteByte(v)
			capNext = true
		} else {
			capNext = v == '_' || v == ' ' || v == '-' || v == '.'
		}
	}
	return n.String()
}
