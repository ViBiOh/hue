package hue

const (
	monday    = 1 << 6
	tuesday   = 1 << 5
	wednesday = 1 << 4
	thursday  = 1 << 3
	friday    = 1 << 2
	saturday  = 1 << 1
	sunday    = 1

	weekday = monday | tuesday | wednesday | thursday | friday
	weekend = saturday | sunday
	alldays = weekday | weekend

	// HueSource constant for worker message
	HueSource = `hue`
)

var (
	// States available states of lights
	States = map[string]map[string]interface{}{
		`off`: {
			`on`:             false,
			`transitiontime`: 30,
		},
		`on`: {
			`on`:             true,
			`transitiontime`: 30,
			`sat`:            0,
			`bri`:            254,
		},
		`dimmed`: {
			`on`:             true,
			`transitiontime`: 30,
			`sat`:            0,
			`bri`:            0,
		},
		`long_on`: {
			`on`:             true,
			`transitiontime`: 3000,
			`sat`:            0,
			`bri`:            254,
		},
	}
)

// Group description
type Group struct {
	Name   string      `json:"name,omitempty"`
	Tap    bool        `json:"tap,omitempty"`
	Lights []string    `json:"lights,omitempty"`
	State  *groupState `json:"state,omitempty"`
}

type groupState struct {
	AnyOn bool `json:"any_on"`
}

// Light description
type Light struct {
	Type  string      `json:"type,omitempty"`
	State *lightState `json:"state,omitempty"`
}

type lightState struct {
	On bool `json:"on,omitempty"`
}

// APIScene describe scene as from Hue API
type APIScene struct {
	Name        string                            `json:"name,omitempty"`
	Lights      []string                          `json:"lights,omitempty"`
	Lightstates map[string]map[string]interface{} `json:"lightstates,omitempty"`
	Recycle     bool                              `json:"recycle"`
}

// Scene description
type Scene struct {
	ID string `json:"id,omitempty"`
	*APIScene
}

// APISchedule describe schedule as from Hue API
type APISchedule struct {
	Name      string  `json:"name,omitempty"`
	Localtime string  `json:"localtime,omitempty"`
	Command   *Action `json:"command,omitempty"`
	Status    string  `json:"status,omitempty"`
}

// Schedule description
type Schedule struct {
	ID string `json:"id,omitempty"`
	*APISchedule
}

// ScheduleConfig configuration (made simple)
type ScheduleConfig struct {
	Name      string
	Localtime string
	Group     string
	State     string
}

// Rule description
type Rule struct {
	ID         string       `json:"-"`
	Status     string       `json:"status,omitempty"`
	Name       string       `json:"name,omitempty"`
	Actions    []*Action    `json:"actions,omitempty"`
	Conditions []*Condition `json:"conditions,omitempty"`
}

// Action description
type Action struct {
	Address string                 `json:"address,omitempty"`
	Body    map[string]interface{} `json:"body,omitempty"`
	Method  string                 `json:"method,omitempty"`
}

// Condition description
type Condition struct {
	Address  string `json:"address,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
}

// Data stores data fo hub
type Data struct {
	Groups    map[string]*Group
	Scenes    map[string]*Scene
	Schedules map[string]*Schedule
	States    map[string]map[string]interface{}
}
