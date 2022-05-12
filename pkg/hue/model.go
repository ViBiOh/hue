package hue

import (
	"regexp"
)

var (
	// States available states of lights
	States = map[string]map[string]any{
		"off": {
			"on":             false,
			"transitiontime": 30,
		},
		"long_off": {
			"on":             false,
			"transitiontime": 300,
		},
		"on": {
			"on":             true,
			"transitiontime": 30,
			"sat":            0,
			"bri":            255,
		},
		"half": {
			"on":             true,
			"transitiontime": 30,
			"sat":            0,
			"bri":            96,
		},
		"dimmed": {
			"on":             true,
			"transitiontime": 30,
			"sat":            0,
			"bri":            0,
		},
		"long_on": {
			"on":             true,
			"transitiontime": 3000,
			"sat":            0,
			"bri":            255,
		},
	}

	scheduleGroupFinder = regexp.MustCompile(`(?mi)groups/(.*?)/`)
)

// Event from the server sent event
type Event struct {
	Type string `json:"type"`
	Data []struct {
		ID   string `json:"id"`
		Type string `json:"type"`

		Temperature struct {
			Temperature float64 `json:"temperature"`
		} `json:"temperature"`

		Light struct {
			Level float64 `json:"light_level"`
		} `json:"light"`

		Motion struct {
			Motion bool `json:"motion"`
		} `json:"motion"`

		On struct {
			On bool `json:"on"`
		} `json:"on"`

		Dimming struct {
			Brightness float64 `json:"brightness"`
		} `json:"dimming"`
	} `json:"data"`
}

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
	Lightstates map[string]map[string]any `json:"lightstates,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Lights      []string                  `json:"lights,omitempty"`
	Recycle     bool                      `json:"recycle"`
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
	Address string         `json:"address,omitempty"`
	Body    map[string]any `json:"body,omitempty"`
	Method  string         `json:"method,omitempty"`
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
