package cron_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/cron"
)

type Fixture struct {
	input    string
	expected string
}

func TestParseCron(t *testing.T) {
	t.Parallel()
	fixtures := []Fixture{
		{
			"every 10 minutes",
			"*/10 * ? * * *",
		},
		{
			"   EveRy    10    mInUteS  ", // not fussy about whitespace or casing
			"*/10 * ? * * *",
		},
		{
			"every 30 minutes from 9am to 5pm",
			"*/30 9-17 ? * * *",
		},
		{
			"every 2 hours",
			"0 */2 ? * * *",
		},
		{
			"every 2 hours from 7am to 7pm",
			"0 7,9,11,13,15,17,19 ? * * *",
		},
		{
			"every monday at 10am",
			"0 10 ? * MON *",
		},
		{
			"every monday,wednesday,friday at 10am",
			"0 10 ? * MON,WED,FRI *",
		},
		{
			"every monday, wednesday, friday at 10am",
			"0 10 ? * MON,WED,FRI *",
		},
		{
			"every monday, wednesday, and friday at 10am",
			"0 10 ? * MON,WED,FRI *",
		},
		{
			"every tuesday and thursday at 9pm",
			"0 21 ? * TUE,THU *",
		},
		{
			"every tuesday and thursday and friday at 9pm",
			"0 21 ? * TUE,THU,FRI *",
		},
		{
			"every day at 4pm",
			"0 16 ? * * *",
		},
		{
			"every weekday at 7am",
			"0 7 ? * MON-FRI *",
		},
		{
			"every weekday at 9am and 5pm",
			"0 9,17 ? * MON-FRI *",
		},
		{
			"every monday,tuesday,wednesday at 9am,12pm,5pm",
			"0 9,12,17 ? * MON,TUE,WED *",
		},
		{
			"every day at 12am",
			"0 0 ? * * *",
		},
		{
			"* * * * *",
			"* * ? * * *", // day-of-month becomes '?'
		},
		{
			"* * * * 1,3,5",         // unix cron uses 0-indexed days-of-week starting on sunday
			"* * ? * MON,WED,FRI *", // day-of-month becomes '?' and convert int values to named
		},
		{
			"* * * 1,4,7,10 *",
			"* * ? JAN,APR,JUL,OCT * *", // day-of-month becomes '?' and convert int values to named
		},
		{
			"*/2 */3 */4 */5 *",   // weird but valid cron - every 2nd minute of every 3rd hour of every 4th day of the month of every 5th month
			"*/2 */3 */4 */5 ? *", // day-of-week becomes '?' because day-of-month has value
		},
		{
			"* * * JAN-MAR MON-FRI",
			"* * ? JAN-MAR MON-FRI *", // day-of-month becomes '?'
		},
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.input, func(t *testing.T) {
			t.Parallel()
			s, err := cron.Parse(fixture.input)
			require.NoError(t, err)
			assert.Equal(t, fixture.expected, s.String())
		})
	}
}

func TestParseCronError(t *testing.T) {
	t.Parallel()
	fixtures := []Fixture{
		{
			"9am on mondays",
			"invalid schedule - must be expression like 'every day at 9am' or cron syntax e.g. '0 9 * * *'",
		},
		{
			"every",
			"unexpected end of input - expected time interval or day e.g. '10 minutes' or 'monday'",
		},
		{
			"every 20",
			"unexpected end of input - expected 'minutes' or 'hours'",
		},
		{
			"every 20 cats",
			"(10:14) unexpected token 'cats' - expected 'minutes' or 'hours'",
		},
		{
			"every 13 minutes",
			"(7:9) value of minutes must divide evenly by 60",
		},
		{
			"every 5 hours",
			"(7:8) value of hours must divide evenly by 24",
		},
		{
			"every 10 minutes from 5pm to 10am",
			"(30:34) end time of schedule must be before start time",
		},
		{
			"every 10 minutes from 5pm until 10am",
			"(27:32) unexpected token 'until' - expected 'to'",
		},
		{
			"every 2 hours between 9am and 3pm",
			"(15:22) unexpected token 'between' - expected 'from'",
		},
		{
			"every 2 hours from breakfast to 3pm",
			"(20:29) invalid time 'breakfast' - must be 12-hour format e.g. '9am'",
		},
		{
			"every 2 hours from 9am to sunset",
			"(27:33) invalid time 'sunset' - must be 12-hour format e.g. '9am'",
		},
		{
			"every funday at 10am",
			"(7:13) invalid day 'funday' - expected day of week e.g. 'monday', 'day' for every day, or 'weekday' for monday-friday",
		},
		{
			"every wednesday before 12pm",
			"(17:23) unexpected token 'before' - expected 'at'",
		},
		{
			"every wednesday at midday",
			"(20:26) invalid time 'midday' - must be 12-hour format e.g. '9am'",
		},
		{
			"every monday at 13am",
			"(17:21) invalid 12-hour time '13am'",
		},
		{
			"every monday at",
			"unexpected end of input - expected time e.g. '9am'",
		},
		{
			"every saturday at 14:00",
			"(19:24) 24-hour format isn't supported",
		},
		{
			"every monday and weekday at 12pm",
			"(18:25) cannot use 'weekday' with other values",
		},
		{
			"every day at 9am on the dot",
			"(18:20) unexpected token 'on' - expected end of input",
		},
		{
			"* * *",
			"wrong number of fields - cron expression must have five fields",
		},
		{
			"*/foo * * * *",
			"(1:6) invalid step value 'foo' for seconds field - must be integer between 1 and 59",
		},
		{
			"*/71 * * * *",
			"(1:5) invalid step value '71' for seconds field - must be integer between 1 and 59",
		},
		{
			"20-10 * * * *",
			"(1:6) invalid range - left value must be smaller than right value",
		},
		{
			"0 42 * * *",
			"(3:5) invalid value '42' for hours field - must be integer between 0 and 23",
		},
		{
			"0 9 35 * *",
			"(5:7) invalid value '35' for day-of-month field - must be integer between 1 and 31",
		},
		{
			"0 9 * 13 *",
			"(7:9) invalid value '13' for month field - must be integer between 1 and 12",
		},
		{
			"0 9 * FOO *",
			"(7:10) invalid value 'FOO' for month field",
		},
		{
			"0 9 * JAN-ERP *",
			"(7:14) invalid value 'ERP' for month field",
		},
		{
			"0 9 * * */2",
			"(9:12) step values are not allowed in day-of-week field",
		},
		{
			"0 9 1 * MON",
			"cannot specify values for both day-of-month and day-of-week - if one is provided the other must be '*'",
		},
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.input, func(t *testing.T) {
			t.Parallel()
			s, err := cron.Parse(fixture.input)
			assert.Nil(t, s)
			require.NotNil(t, err)
			assert.Equal(t, fixture.expected, err.Error())
		})
	}
}
