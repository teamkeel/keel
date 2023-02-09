package value

import (
	"fmt"
	"time"
)

// Value gives a well-defined type to arbitrary values being passed around the keel runtime.
// These may come from, e.g., user input or database query results.
type Value interface {
	// Unexported method to prohibit interface implementation outside of this package.
	// This is important to maintain a closed-set universe of types.
	sealed()
}

func From(value any) (Value, error) {
	switch v := value.(type) {
	case Value:
		return v, nil
	case nil:
		return Nil{}, nil
	case bool:
		return Bool(v), nil
	case string:
		return String(v), nil
	case int:
		return Int(v), nil
	case int64:
		return Int64(v), nil
	case time.Time:
		return Time(v), nil
	default:
		return nil, fmt.Errorf("unsupported Value conversion from %T", value)
	}
}

// Switch allows exhaustive matching over the different values of Value.
// Each "case" function argument is intended to handle one of the cases.
// All error handling should be done outside of it.
//
// Example where we only allow nil or string values:
// ```
//
//	func AssertNilOrString(value Value) (*string, error) {
//		var result *string
//		var err error
//		Switch(
//			value,
//			func() {
//				result = nil
//			},
//			func(value bool) {
//				err = errors.New("expected nil or string but got bool")
//			},
//			func(value string) {
//				result = &value
//			},
//			func(value int) {
//				err = errors.New("expected nil or string but got int")
//			},
//			func(value int64) {
//				err = errors.New("expected nil or string but got int64")
//			},
//			func(value time.Time) {
//				err = errors.New("expected nil or string but got time.Time")
//			},
//		)
//		return result, err
//	}
//
// ```
func Switch(
	value Value,
	caseNil func(),
	caseBool func(value bool),
	caseString func(value string),
	caseInt func(value int),
	caseInt64 func(value int64),
	caseTime func(value time.Time),
) {
	switch v := value.(type) {
	case nil:
		caseNil()
	case Bool:
		caseBool(bool(v))
	case String:
		caseString(string(v))
	case Int:
		caseInt(int(v))
	case Int64:
		caseInt64(int64(v))
	case Time:
		caseTime(time.Time(v))
	}
}

type Nil struct{}
type Bool bool
type String string
type Int int
type Int64 int64
type Time time.Time

func (v Nil) sealed()    {}
func (v Bool) sealed()   {}
func (v String) sealed() {}
func (v Int) sealed()    {}
func (v Int64) sealed()  {}
func (v Time) sealed()   {}
