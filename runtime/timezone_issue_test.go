package runtime

import (
	"testing"
	"time"
)

// TestIllustrateTimeZoneIssue is a temporary test to help us get to the bottom of
// a problem we encountered (at some later date).
//
// It's about time.Unix(). That function is defined to return a local time.
//
// If you run this test, the time object created has a Location field of its nil
// value. This tends to suggest that the go test runner creates a test context
// in which no local time zone is set.
//
// However using the *exact same call - with the same arguments from actions.parseTimeOperand(), in the context of
// TestRuntime() in this package, it returns a time.Time with a one hour offset.
// (It was during GMT) that I ran both tests.
//
// This suggests that somewhere inside our TestRuntime(), we are *setting* a local time zone.
//
// I may have got the reasoning wrong, but one thing I can guarantee is that time.Unix(9,0) returns
// a different thing (in respect of timezone offset) from the two test function entry points, on the
// same computer, when run at the same time. (Which made debugging the problem very hard - and I want sympathy)
func TestIllustrateTimeZoneIssue(t *testing.T) {
	aTime := time.Unix(9, 0)
	_ = aTime
	a := 1
	_ = a
}
