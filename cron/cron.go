package cron

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type Error struct {
	Message string
	Token   *Token
}

func (e Error) Error() string {
	position := ""
	if e.Token != nil {
		position = fmt.Sprintf("(%d:%d) ", e.Token.Start, e.Token.End)
	}
	return fmt.Sprintf("%s%s", position, e.Message)
}

func ToError(err error) (Error, bool) {
	if err == nil {
		return Error{}, false
	}
	e, ok := err.(Error)
	return e, ok
}

func Parse(src string) (*CronExpression, error) {
	if strings.TrimSpace(src) == "" {
		return nil, Error{Message: "input is empty"}
	}

	tokens := toTokens(src)
	token := tokens.Peek()

	// If first token is an integer, integer range, or starts with '*' then parse as unix-cron
	if intValueRegex.MatchString(token.Value) || strings.HasPrefix(token.Value, "*") {
		return parseUnixCron(toTokens(src))
	}

	// Otherwise lower-case input and parse as human-readable expression
	tokens = toTokens(strings.ToLower(src))

	// First token must be 'every'
	token = tokens.Next()
	if token.Value != "every" {
		return nil, Error{Message: "invalid schedule - must be expression like 'every day at 9am' or cron syntax e.g. '0 9 * * *'"}
	}

	token = tokens.Peek()
	if token == nil {
		return nil, Error{Message: "unexpected end of input - expected time interval or day e.g. '10 minutes' or 'monday'"}
	}

	var c *CronExpression

	// If next token is integer then parse as interval expression e.g. 'every 10 minutes'.
	// Else parse as day expression e.g. 'every monday at 9am'
	_, err := strconv.Atoi(token.Value)
	if err == nil {
		c, err = parseIntervalCron(tokens)
	} else {
		c, err = parseDayCron(tokens)
	}

	if err != nil {
		return c, err
	}

	// Check there are no more tokens
	if tokens.Peek() != nil {
		t := tokens.Next()
		return nil, Error{Message: fmt.Sprintf("unexpected token '%s' - expected end of input", t.Value), Token: t}
	}

	return c, err
}

func parseIntervalCron(tokens *Tokens) (*CronExpression, error) {
	c := &CronExpression{
		DayOfMonth: "?",
		Month:      "*",
		DayOfWeek:  "*",
	}

	// At this point we know we have a next token and that it is an integer
	token := tokens.Next()
	intToken := token
	interval, _ := strconv.Atoi(token.Value)

	token, err := tokens.Match("minutes", "hours")
	if err != nil {
		return nil, err
	}

	switch token.Value {
	case "minutes":
		if 60%interval != 0 {
			return nil, Error{Message: "value of minutes must divide evenly by 60", Token: intToken}
		}
		c.Minutes = fmt.Sprintf("*/%d", interval)
		c.Hours = "*"
	case "hours":
		if 24%interval != 0 {
			return nil, Error{Message: "value of hours must divide evenly by 24", Token: intToken}
		}
		c.Minutes = "0"
		c.Hours = fmt.Sprintf("*/%d", interval)
	}

	// optional "from <start-hour> to <end-hour>"
	token, err = tokens.MatchOrNil("from")
	if err != nil {
		return nil, err
	}
	if token != nil {
		start, err := parseTime(tokens.Next())
		if err != nil {
			return nil, err
		}

		_, err = tokens.Match("to")
		if err != nil {
			return nil, err
		}

		endToken := tokens.Next()
		end, err := parseTime(endToken)
		if err != nil {
			return nil, err
		}

		if end < start {
			return nil, Error{Message: "end time of schedule must be before start time", Token: endToken}
		}

		switch c.Hours {
		case "*":
			// If hours is '*' then expression is like 'every 10 minutes from 9am to 11am' in
			// which case we can just set the hours field to 9-11
			c.Hours = fmt.Sprintf("%d-%d", start, end)
		default:
			// Otherwise the expression is like 'every 2 hours from 9am to 3pm, in which case
			// we need to change the hours field to respect both the range and the interval
			// e.g. from '9-15' to '9,11,13,15'
			h := start
			hours := []int{}
			for h <= end {
				hours = append(hours, h)
				h += interval
			}
			c.Hours = strings.Join(
				lo.Map(hours, func(v int, _ int) string {
					return fmt.Sprintf("%d", v)
				}),
				",",
			)
		}
	}

	return c, nil
}

var dayOfWeekMap = map[string]string{
	"sunday":    "SUN",
	"monday":    "MON",
	"tuesday":   "TUE",
	"wednesday": "WED",
	"thursday":  "THU",
	"friday":    "FRI",
	"saturday":  "SAT",
}

func parseDayCron(tokens *Tokens) (*CronExpression, error) {
	c := &CronExpression{
		Minutes:    "0",
		DayOfMonth: "?",
		Month:      "*",
	}

	// days can be a list e.g. 'every monday, wednesday, and friday ...'
	dayTokens := tokens.List()
	days := []string{}

	for _, token := range dayTokens {
		switch {
		case token.Value == "day":
			if len(dayTokens) > 1 {
				return nil, Error{Message: "cannot use 'day' with other values", Token: token}
			}
			c.DayOfWeek = "*"
		case token.Value == "weekday":
			if len(dayTokens) > 1 {
				return nil, Error{Message: "cannot use 'weekday' with other values", Token: token}
			}
			c.DayOfWeek = "MON-FRI"
		default:
			day, ok := dayOfWeekMap[token.Value]
			if !ok {
				msg := fmt.Sprintf("invalid day '%s' - expected day of week e.g. 'monday'", token.Value)
				if len(days) == 0 {
					msg += ", 'day' for every day, or 'weekday' for monday-friday"
				}
				return nil, Error{Message: msg, Token: token}
			}
			if c.DayOfWeek != "" {
				return nil, Error{Message: "cannot use specific days as well as 'day' or 'weekday'", Token: token}
			}
			days = append(days, day)
		}
	}

	if len(days) > 0 {
		c.DayOfWeek = strings.Join(days, ",")
	}

	_, err := tokens.Match("at")
	if err != nil {
		return nil, err
	}

	// hours can also be a list e.g. 'every monday at 9am and 12pm'
	hours := []string{}
	for _, token := range tokens.List() {
		hour, err := parseTime(token)
		if err != nil {
			return nil, err
		}

		hours = append(hours, fmt.Sprintf("%d", hour))
	}
	if len(hours) == 0 {
		return nil, Error{Message: "unexpected end of input - expected time e.g. '9am'"}
	}

	c.Hours = strings.Join(hours, ",")
	return c, nil
}

type CronFieldConfig struct {
	label      string
	min        int
	max        int
	altValues  []string
	stepValues bool
}

var (
	cronFieldConfigs = []CronFieldConfig{
		{"seconds", 0, 59, []string{}, true},
		{"hours", 0, 23, []string{}, true},
		{"day-of-month", 1, 31, []string{}, true},
		{"month", 1, 12, []string{"", "JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}, true},

		// AWS EventBridge cron does not allow '/' (step values) in the day-of-week field (according to their docs)
		{"day-of-week", 0, 6, []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}, false},
	}
	intValueRegex   = regexp.MustCompile(`^(\d+)(?:-(\d+))?$`)
	namedValueRegex = regexp.MustCompile(`^([a-zA-Z]+)(?:-([a-zA-Z]+))?$`)
)

func parseUnixCron(tokens *Tokens) (*CronExpression, error) {
	// group tokens into cron fields, which are separated by whitespace
	groups := [][]*Token{}
	var prev *Token
	token := tokens.Next()
	for token != nil {
		if prev == nil || token.Start-prev.End > 0 {
			// new group
			groups = append(groups, []*Token{token})
		} else {
			// same group
			idx := len(groups) - 1
			groups[idx] = append(groups[idx], token)
		}
		prev = token
		token = tokens.Next()
	}

	if len(groups) != 5 {
		return nil, Error{Message: "wrong number of fields - cron expression must have five fields"}
	}

	c := &CronExpression{}

	for i, group := range groups {
		// remove any commas
		group = lo.Filter(group, func(t *Token, _ int) bool {
			return t.Value != ","
		})

		cfg := cronFieldConfigs[i]
		values := []string{}

		for _, token := range group {
			switch {
			case token.Value == "*":
				values = append(values, "*")
			case strings.HasPrefix(token.Value, "*/"):
				if !cfg.stepValues {
					return nil, Error{Message: fmt.Sprintf("step values are not allowed in %s field", cfg.label), Token: token}
				}
				min := cfg.min
				if min == 0 {
					min = 1
				}
				trimmed := strings.TrimPrefix(token.Value, "*/")
				n, err := strconv.Atoi(trimmed)
				if err != nil || n < min || n > cfg.max {
					return nil, Error{
						Message: fmt.Sprintf("invalid step value '%s' for %s field - must be integer between %d and %d", trimmed, cfg.label, min, cfg.max),
						Token:   token,
					}
				}
				values = append(values, token.Value)
			case intValueRegex.MatchString(token.Value):
				matches := intValueRegex.FindStringSubmatch(token.Value)
				matches = matches[1:]
				if matches[1] == "" {
					matches = matches[:1]
				}
				ints := lo.Map(matches, func(s string, _ int) int {
					v, _ := strconv.Atoi(s)
					return v
				})
				for i, v := range ints {
					if v < cfg.min || v > cfg.max {
						return nil, Error{
							Message: fmt.Sprintf("invalid value '%d' for %s field - must be integer between %d and %d", v, cfg.label, cfg.min, cfg.max),
							Token:   token,
						}
					}
					if i == 1 && v < ints[0] {
						return nil, Error{
							Message: "invalid range - left value must be smaller than right value",
							Token:   token,
						}
					}
				}
				if len(cfg.altValues) > 0 {
					altValues := lo.Map(ints, func(v int, _ int) string {
						return cfg.altValues[v]
					})
					values = append(values, strings.Join(altValues, "-"))
				} else {
					values = append(values, token.Value)
				}
			case namedValueRegex.MatchString(token.Value):
				matches := namedValueRegex.FindStringSubmatch(token.Value)
				matches = matches[1:]
				if matches[1] == "" {
					matches = matches[:1]
				}
				for _, v := range matches {
					if !lo.Contains(cfg.altValues, strings.ToUpper(v)) {
						return nil, Error{
							Message: fmt.Sprintf("invalid value '%s' for %s field", v, cfg.label),
							Token:   token,
						}
					}
				}
				values = append(values, strings.ToUpper(token.Value))
			default:
				return nil, Error{
					Message: fmt.Sprintf("invalid value '%s' for %s field", token.Value, cfg.label),
					Token:   token,
				}
			}
		}

		fieldValue := strings.Join(values, ",")
		switch i {
		case 0:
			c.Minutes = fieldValue
		case 1:
			c.Hours = fieldValue
		case 2:
			c.DayOfMonth = fieldValue
		case 3:
			c.Month = fieldValue
		case 4:
			c.DayOfWeek = fieldValue
		}
	}

	// From AWS EventBridge cron docs:
	// > You can't specify the Day-of-month and Day-of-week fields in the same cron expression.
	// > If you specify a value or a * (asterisk) in one of the fields, you must use a ? (question mark) in the other.
	switch {
	case c.DayOfMonth == "*":
		c.DayOfMonth = "?"
	case c.DayOfWeek == "*":
		c.DayOfWeek = "?"
	}
	if c.DayOfMonth != "?" && c.DayOfWeek != "?" {
		return nil, Error{Message: "cannot specify values for both day-of-month and day-of-week - if one is provided the other must be '*'"}
	}

	return c, nil
}

var (
	twelveHourTime     = regexp.MustCompile(`^(\d{1,2})(am|pm)$`)
	twentyFourHourTime = regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
)

func parseTime(token *Token) (int, error) {
	if token == nil {
		return 0, Error{Message: "unexpected end of input - expected time e.g. '9am'"}
	}

	switch {
	case twelveHourTime.MatchString(token.Value):
		matches := twelveHourTime.FindStringSubmatch(strings.ToLower(token.Value))
		// Due to the regex we know matches[1] is an integer
		hour, _ := strconv.Atoi(matches[1])

		// But it may be out-of-bounds for a 12-hour time
		if hour < 1 || hour > 12 {
			return 0, Error{Message: fmt.Sprintf("invalid 12-hour time '%s'", token.Value), Token: token}
		}

		if matches[2] == "pm" && hour != 12 {
			hour += 12
		}
		if matches[2] == "am" && hour == 12 {
			hour = 0
		}

		return hour, nil
	case twentyFourHourTime.MatchString(token.Value):
		return 0, Error{Message: "24-hour format isn't supported", Token: token}
	default:
		return 0, Error{
			Message: fmt.Sprintf("invalid time '%s' - must be 12-hour format e.g. '9am'", token.Value),
			Token:   token,
		}
	}
}

type CronExpression struct {
	Minutes    string
	Hours      string
	DayOfMonth string
	Month      string
	DayOfWeek  string
}

func (c *CronExpression) String() string {
	return fmt.Sprintf("%s %s %s %s %s *", c.Minutes, c.Hours, c.DayOfMonth, c.Month, c.DayOfWeek)
}

type Tokens struct {
	tokens    []*Token
	currIndex int
}

// Peek returns the next token or nil but does not advance the current token
func (t *Tokens) Peek() *Token {
	i := t.currIndex
	i++
	if i > len(t.tokens)-1 {
		return nil
	}
	return t.tokens[i]
}

// Next returns the next token or nil and advances the current token
func (t *Tokens) Next() *Token {
	t.currIndex++
	if t.currIndex > len(t.tokens)-1 {
		return nil
	}
	return t.tokens[t.currIndex]
}

// MatchOrNil is like Next() but checks that the next token (if it exists)
// matches one of the expected values
func (t *Tokens) MatchOrNil(expected ...string) (*Token, error) {
	tok := t.Next()
	if tok == nil {
		return nil, nil
	}
	if len(expected) > 0 {
		match := false
		for _, v := range expected {
			if tok.Value == v {
				match = true
			}
		}
		if !match {
			return nil, Error{
				Message: fmt.Sprintf("unexpected token '%s' - expected %s", tok.Value, expectedToString(expected...)),
				Token:   tok,
			}
		}
	}
	return tok, nil
}

// Match is like MatchOrNil but will return an error if the returned token is nil
func (t *Tokens) Match(expected ...string) (*Token, error) {
	tok, err := t.MatchOrNil(expected...)
	if err != nil {
		return tok, err
	}
	if tok == nil {
		return nil, Error{Message: fmt.Sprintf("unexpected end of input - expected %s", expectedToString(expected...))}
	}
	return tok, nil
}

// List starts at the next token and will keep consuming tokens while they are in a list. A list is separated by
// commas and 'and'.
// As an example both 'a,b,c' and 'a, b, and c' would result in the list ['a', 'b', 'c'].
func (t *Tokens) List() []*Token {
	list := []*Token{}
	next := t.Peek()
	isSeperator := true

	for {
		if next == nil {
			return list
		}

		if next.Value == "," || next.Value == "and" {
			_ = t.Next()
			next = t.Peek()
			isSeperator = true
			continue
		}

		if !isSeperator {
			return list
		}

		isSeperator = false
		list = append(list, next)
		_ = t.Next()
		next = t.Peek()
	}
}

func expectedToString(expected ...string) string {
	quoted := lo.Map(expected, func(s string, _ int) string {
		return fmt.Sprintf("'%s'", s)
	})
	switch len(quoted) {
	case 1:
		return quoted[0]
	case 2:
		return strings.Join(quoted, " or ")
	default:
		return strings.Join(quoted[:len(quoted)-1], ", ") + " or " + quoted[len(quoted)-1]
	}
}

type Token struct {
	Value string
	Start int
	End   int
}

// toTokens breaks src up into tokens. It is a very crude/simple lexer.
func toTokens(src string) *Tokens {
	tokens := &Tokens{currIndex: -1}
	var t *Token

	for i, char := range src {
		// Ignore whitespace
		if char == ' ' {
			continue
		}

		// If no token make a new one
		if t == nil {
			t = &Token{
				Start: i + 1,
			}
			tokens.tokens = append(tokens.tokens, t)
		}

		t.Value = fmt.Sprintf("%s%c", t.Value, char)

		// We've reached the end of the current token if one of the following is true:
		// [1] this is the last character of src
		// [2] the next character is whitespace
		// [3] the next character is a comma
		// [4] this character is a comma
		isEndOfToken :=
			(i == len(src)-1 || // [1]
				src[i+1] == ' ' || // [2]
				src[i+1] == ',' || // [3]
				char == ',') // [4]

		if isEndOfToken {
			// +2 here because we want the end position to point to the character
			// after the final character of this token
			t.End = i + 2
			t = nil
		}
	}

	return tokens
}
