package timeperiod

import (
	reflect "reflect"
	"testing"
)

func Test_parseString(t *testing.T) {
	tests := []struct {
		expr    string
		want    TimePeriod
		wantErr bool
	}{
		{
			expr: "this month",
			want: TimePeriod{
				Period:   "month",
				Offset:   0,
				Value:    1,
				Complete: true,
			},
		},
		{
			expr: "this year",
			want: TimePeriod{
				Period:   "year",
				Offset:   0,
				Value:    1,
				Complete: true,
			},
		},
		{
			expr: "next complete day",
			want: TimePeriod{
				Period:   "day",
				Offset:   1,
				Value:    1,
				Complete: true,
			},
		},
		{
			expr: "next day",
			want: TimePeriod{
				Period:   "day",
				Offset:   0,
				Value:    1,
				Complete: false,
			},
		},
		{
			expr: "last day",
			want: TimePeriod{
				Period:   "day",
				Offset:   -1,
				Value:    1,
				Complete: false,
			},
		},
		{
			expr: "last complete day",
			want: TimePeriod{
				Period:   "day",
				Offset:   -1,
				Value:    1,
				Complete: true,
			},
		},
		{
			expr: "last 2 years",
			want: TimePeriod{
				Period:   "year",
				Offset:   -2,
				Value:    2,
				Complete: false,
			},
		},
		{
			expr: "last 2 complete months",
			want: TimePeriod{
				Period:   "month",
				Offset:   -2,
				Value:    2,
				Complete: true,
			},
		},
		{
			expr: "next 5 complete months",
			want: TimePeriod{
				Period:   "month",
				Offset:   1,
				Value:    5,
				Complete: true,
			},
		},
		{
			expr: "next 5 months",
			want: TimePeriod{
				Period:   "month",
				Offset:   0,
				Value:    5,
				Complete: false,
			},
		},
		{
			expr:    "this complete day",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := parseString(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() = %v, want %v", got, tt.want)
			}
		})
	}
}
