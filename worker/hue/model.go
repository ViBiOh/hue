package hue

import "github.com/ViBiOh/iot/hue"

type hueConfig struct {
	Schedules []*scheduleConfig
	Taps      []*tapConfig
}

type scheduleConfig struct {
	Name      string
	Localtime string
	Group     string
	State     string
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
