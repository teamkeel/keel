package types

import (
	"database/sql/driver"
	"time"
)

type Date struct {
	time.Time
}

func (d Date) Value() (driver.Value, error) {
	return d.Time, nil
}

type Timestamp struct {
	time.Time
}

func (t Timestamp) Value() (driver.Value, error) {
	return t.Time, nil
}

type Duration struct {
	Duration string
}

func (t Duration) Value() (driver.Value, error) {
	return t.Duration, nil
}
