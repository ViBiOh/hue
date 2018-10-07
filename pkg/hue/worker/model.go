package hue

import "github.com/ViBiOh/iot/pkg/hue"

type hueConfig struct {
	Schedules []*hue.ScheduleConfig
	Taps      []*tapConfig
}

type tapConfig struct {
	ID      string
	Buttons []*tapButton
}

type tapButton struct {
	ID     string
	State  string
	Groups []string
	Rule   *hue.Rule
}
