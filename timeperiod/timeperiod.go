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

// IsTimezoneRelative determines if the user's timezone would have an impact for this time period.
//
// When handling a date relative period, such as `last complete day`/`yesterday`, this will have an impact:
// Given a current time of 15/11/2024 01:15 in UTC
// * for America/Toronto UTC -5: date_trunc('day', NOW()) should be 14/11/2024
// * for Europe/Rome UTC +1:  date_trunc('day', NOW()) should be 15/11/2024
func (tp *TimePeriod) IsTimezoneRelative() bool {
	if !tp.Complete {
		return false
	}

	switch tp.Period {
	case "day", "week", "month", "year":
		return true
	}

	return false
}

// Parse will return a TimePeriodStruct from the given expression. Expressions follow the following form
//   - {this/last/next} {n}? {complete}? {second(s)/minute(s)/hour(s)/day(s)/week(s)/month(s)/year(s)}
//   - or one of the supported shorthand values: now/today/tomorrow/yesterday
//
// Expression rules:
// * if `this`, then there is no n and no complete; i.e. `this month`, `this day`, `this year`
// * n = positive integer
// * `complete“ is optional
// * period can be either plural or singular version
func Parse(expression string) (TimePeriod, error) {
	tkns := toTokens(expression)

	if err := tkns.validate(); err != nil {
		return TimePeriod{}, fmt.Errorf("invalid time period expression: %w", err)
	}

	if tkns.isShorthand() {
		switch tkns[0] {
		case "now":
			return TimePeriod{}, nil
		case "today":
			return Parse("this day")
		case "tomorrow":
			return Parse("next complete day")
		case "yesterday":
			return Parse("last complete day")
		}
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
	p := strings.TrimSuffix(strings.ToLower(t[len(t)-1]), "s")
	switch p {
	case "second", "minute", "hour", "day", "week", "month", "year":
		return p
	}
	return ""
}

func (t tokens) direction() string {
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
	if t.direction() == "this" {
		return true
	}
	if len(t) < 3 {
		return false
	}

	return strings.ToLower(t[len(t)-2]) == "complete"
}

func (t tokens) value() int {
	switch t.direction() {
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
	switch t.direction() {
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

func (t tokens) isShorthand() bool {
	if len(t) != 1 {
		return false
	}

	switch t[0] {
	case "now", "today", "tomorrow", "yesterday":
		return true
	}

	return false
}

func (t tokens) validate() error {
	if t.isShorthand() {
		return nil
	}

	a := t.direction()
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
		if t.value() < 1 {
			return fmt.Errorf("time period expression should have a positive amount of periods; e.g, `next 5 days")
		}
	default:
		return fmt.Errorf("time period expression should start with this/next/last")
	}
	return nil
}
