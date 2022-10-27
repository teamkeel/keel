package runtimectx

import (
	"time"
)

const (
	nowContextKey contextKey = "now"
)

func GetNow() time.Time {

	return time.Now().UTC()
}
