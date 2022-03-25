package hue

import (
	"testing"
)

func TestFormatLocalTime(t *testing.T) {
	type args struct {
		instance Schedule
	}

	cases := map[string]struct {
		args args
		want string
	}{
		"no prefix": {
			args{
				instance: Schedule{
					ID: "",
					APISchedule: APISchedule{
						Localtime: "08:00:00",
					},
				},
			},
			"08:00:00",
		},
		"invalid number": {
			args{
				instance: Schedule{
					ID: "",
					APISchedule: APISchedule{
						Localtime: "WABC 08:00:00",
					},
				},
			},
			"WABC 08:00:00",
		},
		"all days": {
			args{
				instance: Schedule{
					ID: "",
					APISchedule: APISchedule{
						Localtime: "W127  08:00:00",
					},
				},
			},
			"All days at 08:00:00",
		},
		"week days": {
			args{
				instance: Schedule{
					ID: "",
					APISchedule: APISchedule{
						Localtime: "W124  10:00:00",
					},
				},
			},
			"Week days at 10:00:00",
		},
		"weekend": {
			args{
				instance: Schedule{
					ID: "",
					APISchedule: APISchedule{
						Localtime: "W003  10:00:00",
					},
				},
			},
			"Weekend at 10:00:00",
		},
		"mon, wed, fri, sun": {
			args{
				instance: Schedule{
					ID: "",
					APISchedule: APISchedule{
						Localtime: "W085  12:00:00",
					},
				},
			},
			"Mon, Wed, Fri, Sun at 12:00:00",
		},
		"tue, thu, sat": {
			args{
				instance: Schedule{
					ID: "",
					APISchedule: APISchedule{
						Localtime: "W042  14:00:00",
					},
				},
			},
			"Tue, Thu, Sat at 14:00:00",
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := tc.args.instance.FormatLocalTime(); got != tc.want {
				t.Errorf("FormatLocalTime() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}
