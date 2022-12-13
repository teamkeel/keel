package actions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

func toNative(v *parser.Operand, fieldType proto.Type) (any, error) {
	switch {
	case v.False:
		return false, nil
	case v.True:
		return true, nil
	case v.Number != nil:
		return *v.Number, nil
	case v.String != nil:
		v := *v.String
		v = strings.TrimPrefix(v, `"`)
		v = strings.TrimSuffix(v, `"`)
		switch fieldType {
		case proto.Type_TYPE_DATE:
			return toDate(v), nil
		case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
			return toTime(v), nil
		}
		return v, nil
	case v.Null:
		return nil, nil
	case fieldType == proto.Type_TYPE_ENUM:
		return v.Ident.Fragments[0].Fragment, nil
	default:
		return nil, fmt.Errorf("toNative() does yet support this expression operand: %+v", v)
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
