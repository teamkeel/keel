package timeperiod

import (
	"fmt"
	"strconv"
	"strings"
)

type TimePeriod struct {
	Period   string
	Value    int
	Offset   int
	Complete bool
}

// Parse will return a TimePeriodStruct from the given expression. Expressions follow the following form
// {this/last/next} {n}? {full}? {hour/day/week/month/year}
//
// if `this`, then there is no n and no full; i.e.
// - this month, this day, this hour
func Parse(expression any) (TimePeriod, error) {
	switch t := expression.(type) {
	case string:
		return parseString(t)
	}

	return TimePeriod{}, fmt.Errorf("unsupported expression type")
}

func parseString(expr string) (TimePeriod, error) {
	tkns := toTokens(expr)

	if err := tkns.validate(); err != nil {
		return TimePeriod{}, fmt.Errorf("invalid time period expression: %w", err)
	}

	return TimePeriod{
		Period:   tkns.period(),
		Offset:   tkns.offset(),
		Value:    tkns.value(),
		Complete: tkns.complete(),
	}, nil
}

type tokens []string

func toTokens(expr string) tokens {
	t := strings.Fields(expr)
	return tokens(t)
}

// period will extract the period from the tokens, will return "" if no valid period. The period should always be at the
// end of the expression
func (t tokens) period() string {
	p := strings.ToLower(strings.TrimSuffix(t[len(t)-1], "s"))
	switch p {
	case "hour", "day", "week", "month", "year":
		return p
	}
	return ""
}

func (t tokens) attribute() string {
	if len(t) == 0 {
		return ""
	}
	switch strings.ToLower(t[0]) {
	case "this":
		return "this"
	case "next":
		return "next"
	case "last":
		return "last"
	}

	return ""
}

// complete
func (t tokens) complete() bool {
	if t.attribute() == "this" {
		return true
	}
	if len(t) < 3 {
		return false
	}

	return strings.ToLower(t[len(t)-2]) == "complete"
}

func (t tokens) value() int {
	switch t.attribute() {
	case "this":
		return 1
	case "next", "last":
		if v, err := strconv.ParseInt(t[1], 10, 32); err == nil {
			return int(v)
		}
	}
	return 1
}

func (t tokens) offset() int {
	switch t.attribute() {
	case "next":
		if t.complete() {
			return 1
		}
		return 0
	case "last":
		return -t.value()
	}
	return 0
}

func (t tokens) validate() error {
	a := t.attribute()
	if a == "" {
		return fmt.Errorf("time period expression should start with this/next/last")
	}

	switch a {
	case "this":
		if len(t) != 2 || t.period() == "" {
			return fmt.Errorf("time period expression should be in the form of `this {day/week/month/year}`")
		}
	case "next", "last":
		if t.period() == "" {
			return fmt.Errorf("time period expression should have a valid period; e.g. day/week/month/year`")
		}
		if len(t) > 4 || (len(t) == 4 && !t.complete()) {
			return fmt.Errorf("time period expression should be in the form of `{next/last} {n}? {complete}? {day/week/month/year}`")
		}
	default:
		return fmt.Errorf("time period expression should start with this/next/last")
	}
	return nil
}
