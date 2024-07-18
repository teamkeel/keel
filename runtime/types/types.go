package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Date struct {
	time.Time
}

func (d Date) Value() (driver.Value, error) {
	return d.Time, nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(time.DateOnly))
}

type Timestamp struct {
	time.Time
}

func (t Timestamp) Value() (driver.Value, error) {
	return t.Time, nil
}
