package hue

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

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
	Name    string   `json:"name,omitempty"`
	Lights  []string `json:"lights,omitempty"`
	Recycle bool     `json:"recycle"`
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

// FormatLocalTime formats local time of schedules to human readable version
func (s *Schedule) FormatLocalTime() string {
	if !strings.HasPrefix(s.Localtime, `W`) {
		return s.Localtime
	}

	recurrence, err := strconv.Atoi(s.Localtime[1:4])
	if err != nil {
		log.Printf(`Error while parsing local time: %v`, err)
		return s.Localtime
	}

	var recurrenceStr string

	if recurrence == alldays {
		recurrenceStr = `All days`
	} else if recurrence == weekday {
		recurrenceStr = `Week days`
	} else if recurrence == weekend {
		recurrenceStr = `Weekend`
	} else {
		days := make([]string, 5)

		if recurrence&monday != 0 {
			days = append(days, `Mon`)
		}
		if recurrence&tuesday != 0 {
			days = append(days, `Tue`)
		}
		if recurrence&wednesday != 0 {
			days = append(days, `Wed`)
		}
		if recurrence&thursday != 0 {
			days = append(days, `Thu`)
		}
		if recurrence&friday != 0 {
			days = append(days, `Fri`)
		}
		if recurrence&saturday != 0 {
			days = append(days, `Sat`)
		}
		if recurrence&sunday != 0 {
			days = append(days, `Sun`)
		}

		recurrenceStr = strings.Join(days, `, `)
	}

	return fmt.Sprintf(`%s at %s`, recurrenceStr, s.Localtime[6:])
}

// ComputeScheduleReccurence formats local time of schedules to API version
func ComputeScheduleReccurence(days []string, hours, minutes string) string {
	var recurrence int

	for _, day := range days {
		if day == `Mon` {
			recurrence |= monday
		}
		if day == `Tue` {
			recurrence |= tuesday
		}
		if day == `Wed` {
			recurrence |= wednesday
		}
		if day == `Thu` {
			recurrence |= thursday
		}
		if day == `Fri` {
			recurrence |= friday
		}
		if day == `Sat` {
			recurrence |= saturday
		}
		if day == `Sun` {
			recurrence |= sunday
		}
	}

	return fmt.Sprintf(`W%dT%s:%s:00`, recurrence, hours, minutes)
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
