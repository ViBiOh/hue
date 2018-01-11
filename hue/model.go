package hue

// Group description
type Group struct {
	Name   string      `json:"name"`
	Tap    bool        `json:"tap"`
	Lights []string    `json:"lights"`
	State  *groupState `json:"state"`
}

type groupState struct {
	AnyOn bool `json:"any_on"`
}

// Light description
type Light struct {
	Type  string      `json:"type"`
	State *lightState `json:"state"`
}

type lightState struct {
	On bool `json:"on"`
}

// Scene description
type Scene struct {
	ID     string   `json:"-"`
	Name   string   `json:"name"`
	Lights []string `json:"lights"`
}

// Schedule description
type Schedule struct {
	ID        string   `json:"-"`
	Name      string   `json:"name"`
	Localtime string   `json:"localtime"`
	Lights    []string `json:"lights"`
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
