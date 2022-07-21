package actions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/expressions"
)

func toNative(v *expressions.Operand, fieldType proto.Type) any {
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
		switch fieldType {
		case proto.Type_TYPE_DATE:
			return toDate(v)
		case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
			return toTime(v)
		}
		return v
	default:
		panic(fmt.Sprintf("toNative() does yet support this expression operand: %+v", v))
	}
}

func toDate(s string) time.Time {
	segments := strings.Split(s, `/`)
	day, _ := strconv.Atoi(segments[0])
	month, _ := strconv.Atoi(segments[1])
	year, _ := strconv.Atoi(segments[2])
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func toTime(s string) time.Time {
	tm, _ := time.Parse(time.RFC3339, s)
	return tm
}
