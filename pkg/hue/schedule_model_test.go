package hue

import (
	"testing"
)

func TestScheduleTime(t *testing.T) {
	cases := map[string]struct {
		instance Schedule
		want     string
	}{
		"valid time with recurrence": {
			instance: Schedule{
				APISchedule: APISchedule{
					Localtime: "W127/T08:00:00",
				},
			},
			want: "08:00",
		},
		"valid time without recurrence": {
			instance: Schedule{
				APISchedule: APISchedule{
					Localtime: "W000/T14:30:00",
				},
			},
			want: "14:30",
		},
		"invalid format": {
			instance: Schedule{
				APISchedule: APISchedule{
					Localtime: "08:00:00",
				},
			},
			want: "",
		},
		"empty string": {
			instance: Schedule{
				APISchedule: APISchedule{
					Localtime: "",
				},
			},
			want: "",
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := tc.instance.ScheduleTime(); got != tc.want {
				t.Errorf("ScheduleTime() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestHasDay(t *testing.T) {
	type args struct {
		instance Schedule
		dayValue int
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"all days - monday": {
			args{
				instance: Schedule{
					APISchedule: APISchedule{
						Localtime: "W127  08:00:00",
					},
				},
				dayValue: monday,
			},
			true,
		},
		"all days - tuesday": {
			args{
				instance: Schedule{
					APISchedule: APISchedule{
						Localtime: "W127  08:00:00",
					},
				},
				dayValue: tuesday,
			},
			true,
		},
		"week days - monday": {
			args{
				instance: Schedule{
					APISchedule: APISchedule{
						Localtime: "W124  10:00:00",
					},
				},
				dayValue: monday,
			},
			true,
		},
		"week days - saturday": {
			args{
				instance: Schedule{
					APISchedule: APISchedule{
						Localtime: "W124  10:00:00",
					},
				},
				dayValue: saturday,
			},
			false,
		},
		"weekend - sunday": {
			args{
				instance: Schedule{
					APISchedule: APISchedule{
						Localtime: "W003  10:00:00",
					},
				},
				dayValue: sunday,
			},
			true,
		},
		"weekend - monday": {
			args{
				instance: Schedule{
					APISchedule: APISchedule{
						Localtime: "W003  10:00:00",
					},
				},
				dayValue: monday,
			},
			false,
		},
		"mon, wed, fri, sun - monday": {
			args{
				instance: Schedule{
					APISchedule: APISchedule{
						Localtime: "W085  12:00:00",
					},
				},
				dayValue: monday,
			},
			true,
		},
		"mon, wed, fri, sun - tuesday": {
			args{
				instance: Schedule{
					APISchedule: APISchedule{
						Localtime: "W085  12:00:00",
					},
				},
				dayValue: tuesday,
			},
			false,
		},
		"no prefix - any day": {
			args{
				instance: Schedule{
					APISchedule: APISchedule{
						Localtime: "08:00:00",
					},
				},
				dayValue: monday,
			},
			false,
		},
		"invalid recurrence - any day": {
			args{
				instance: Schedule{
					APISchedule: APISchedule{
						Localtime: "WABC  08:00:00",
					},
				},
				dayValue: monday,
			},
			false,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := tc.args.instance.HasDay(tc.args.dayValue); got != tc.want {
				t.Errorf("HasDay() = `%t`, want `%t`", got, tc.want)
			}
		})
	}
}

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
