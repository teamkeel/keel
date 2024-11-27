package timeperiod

import (
	reflect "reflect"
	"testing"
)

func Test_Parse(t *testing.T) {
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
			expr: "this    year  ",
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
			expr: "  next day",
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
			expr: "next 5 MONTHS",
			want: TimePeriod{
				Period:   "month",
				Offset:   0,
				Value:    5,
				Complete: false,
			},
		},
		{
			expr: "now",
			want: TimePeriod{
				Period:   "",
				Offset:   0,
				Value:    0,
				Complete: false,
			},
		},
		{
			expr: "today",
			want: TimePeriod{
				Period:   "day",
				Offset:   0,
				Value:    1,
				Complete: true,
			},
		},
		{
			expr: "tomorrow",
			want: TimePeriod{
				Period:   "day",
				Offset:   1,
				Value:    1,
				Complete: true,
			},
		},
		{
			expr: "yesterday     ",
			want: TimePeriod{
				Period:   "day",
				Offset:   -1,
				Value:    1,
				Complete: true,
			},
		},
		{
			expr:    "5 days",
			wantErr: true,
		},
		{
			expr:    "this  complete day",
			wantErr: true,
		},
		{
			expr:    "next something wrong",
			wantErr: true,
		},
		{
			expr:    "next 0 days",
			wantErr: true,
		},
		{
			expr:    "next -10 months",
			wantErr: true,
		},
		{
			expr:    "today 3 months",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := Parse(tt.expr)
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

func TestTimePeriod_IsTimezoneRelative(t *testing.T) {
	tests := []struct {
		expr string
		want bool
	}{
		{
			expr: "this month",
			want: true,
		},
		{
			expr: "this year",
			want: true,
		},
		{
			expr: "this hour",
			want: false,
		},
		{
			expr: "yesterday",
			want: true,
		},
		{
			expr: "last month",
			want: false,
		},
		{
			expr: "next complete week",
			want: true,
		},
		{
			expr: "next 5 complete hours",
			want: false,
		},
		{
			expr: "next 5 complete days",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			tp, err := Parse(tt.expr)
			if err != nil {
				t.Error("failed parsing expression: %w", err)
			}
			if got := tp.IsTimezoneRelative(); got != tt.want {
				t.Errorf("TimePeriod.IsTimezoneRelative() %s => %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}
