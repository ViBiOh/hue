package hue

import (
	"regexp"
	"time"
)

// State description
type State struct {
	On         bool          `json:"on"`
	Duration   time.Duration `json:"transitiontime"`
	Saturation uint64        `json:"sat"`
	Brightness uint64        `json:"bri"`
}

var (
	// States available states of lights
	States = map[string]State{
		"off": {
			On:       false,
			Duration: time.Second * 3,
		},
		"long_off": {
			On:       false,
			Duration: 300,
		},
		"on": {
			On:         true,
			Duration:   time.Second * 3,
			Saturation: 0,
			Brightness: 100,
		},
		"half": {
			On:         true,
			Duration:   time.Second * 3,
			Saturation: 0,
			Brightness: 96,
		},
		"dimmed": {
			On:         true,
			Duration:   time.Second * 3,
			Saturation: 0,
			Brightness: 0,
		},
		"long_on": {
			On:         true,
			Duration:   time.Minute * 5,
			Saturation: 0,
			Brightness: 100,
		},
	}

	scheduleGroupFinder = regexp.MustCompile(`(?mi)groups/(.*?)/`)
)

// Group description
type Group struct {
	ID     string     `json:"id,omitempty"`
	Name   string     `json:"name,omitempty"`
	Lights []string   `json:"lights,omitempty"`
	State  groupState `json:"state,omitempty"`
	Tap    bool       `json:"tap,omitempty"`
}

type groupState struct {
	AnyOn bool `json:"any_on"`
}

// Light description
type Light struct {
	ID    string     `json:"id,omitempty"`
	Type  string     `json:"type,omitempty"`
	State lightState `json:"state,omitempty"`
}

type lightState struct {
	On bool `json:"on,omitempty"`
}

// APIScene describe scene as from Hue API
type APIScene struct {
	Lightstates map[string]State `json:"lightstates,omitempty"`
	Name        string           `json:"name,omitempty"`
	Lights      []string         `json:"lights,omitempty"`
	Recycle     bool             `json:"recycle"`
}

// Scene description
type Scene struct {
	ID string `json:"id,omitempty"`
	APIScene
}

// Rule description
type Rule struct {
	ID         string      `json:"-"`
	Status     string      `json:"status,omitempty"`
	Name       string      `json:"name,omitempty"`
	Actions    []Action    `json:"actions,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

// Sensor description
type Sensor struct {
	ID     string       `json:"-"`
	Name   string       `json:"name,omitempty"`
	Type   string       `json:"type,omitempty"`
	State  sensorState  `json:"state,omitempty"`
	Config SensorConfig `json:"config,omitempty"`
}

// BySensorID sort Sensor by id
type BySensorID []Sensor

func (a BySensorID) Len() int      { return len(a) }
func (a BySensorID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySensorID) Less(i, j int) bool {
	return a[i].ID < a[j].ID
}

type sensorState struct {
	Presence    bool    `json:"presence"`
	Temperature float32 `json:"temperature,omitempty"`
}

// SensorConfig description
type SensorConfig struct {
	Battery       uint `json:"battery,omitempty"`
	On            bool `json:"on"`
	LedIndication bool `json:"ledindication"`
}

// Action description
type Action struct {
	Address string `json:"address,omitempty"`
	Body    any    `json:"body,omitempty"`
	Method  string `json:"method,omitempty"`
}

// GetGroup returns the group ID of the Action performed
func (a Action) GetGroup() string {
	matches := scheduleGroupFinder.FindStringSubmatch(a.Address)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// Condition description
type Condition struct {
	Address  string `json:"address,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
}
