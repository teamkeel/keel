package actions

import (
	"fmt"
	"strings"

	"github.com/teamkeel/keel/schema/expressions"
)

func toNative(v *expressions.Operand) any {
	switch {
	case v.False:
		return false
	case v.True:
		return true
	case v.Number != nil:
		return *v.Number
	case v.String != nil:
		v := *v.String
		v = strings.TrimPrefix(v, `"`)
		v = strings.TrimSuffix(v, `"`)
		return v
	default:
		panic(fmt.Sprintf("toNative() does yet support this expression operand: %+v", v))
	}
}
