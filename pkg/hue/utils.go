package hue

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
)

// FormatLocalTime formats local time of schedules to human readable version
func (s *Schedule) FormatLocalTime() string {
	if !strings.HasPrefix(s.Localtime, "W") {
		return s.Localtime
	}

	recurrence, err := strconv.Atoi(s.Localtime[1:4])
	if err != nil {
		logger.Error("%#v", errors.WithStack(err))
		return s.Localtime
	}

	var recurrenceStr string

	if recurrence == alldays {
		recurrenceStr = "All days"
	} else if recurrence == weekday {
		recurrenceStr = "Week days"
	} else if recurrence == weekend {
		recurrenceStr = "Weekend"
	} else {
		days := make([]string, 0)

		if recurrence&monday != 0 {
			days = append(days, "Mon")
		}
		if recurrence&tuesday != 0 {
			days = append(days, "Tue")
		}
		if recurrence&wednesday != 0 {
			days = append(days, "Wed")
		}
		if recurrence&thursday != 0 {
			days = append(days, "Thu")
		}
		if recurrence&friday != 0 {
			days = append(days, "Fri")
		}
		if recurrence&saturday != 0 {
			days = append(days, "Sat")
		}
		if recurrence&sunday != 0 {
			days = append(days, "Sun")
		}

		recurrenceStr = strings.Join(days, ", ")
	}

	return fmt.Sprintf("%s at %s", recurrenceStr, s.Localtime[6:])
}

func formatStateValue(state map[string]interface{}) string {
	return fmt.Sprintf("%v|%v|%v|%v", state["on"], state["transitiontime"], state["sat"], state["bri"])
}

// FindStateName finds matching state's name
func (s *Schedule) FindStateName(scenes map[string]*Scene) string {
	if sceneID, ok := s.Command.Body["scene"]; ok {
		if scene, ok := scenes[sceneID.(string)]; ok {
			for _, lightState := range scene.Lightstates {
				lightStateValue := formatStateValue(lightState)

				for stateName, state := range States {
					if lightStateValue == formatStateValue(state) {
						return stateName
					}
				}
			}
		}
	}

	return "unknown"
}

// ComputeScheduleReccurence formats local time of schedules to API version
func ComputeScheduleReccurence(days []string, hours, minutes string) string {
	var recurrence int

	for _, day := range days {
		if day == "Mon" {
			recurrence |= monday
		}
		if day == "Tue" {
			recurrence |= tuesday
		}
		if day == "Wed" {
			recurrence |= wednesday
		}
		if day == "Thu" {
			recurrence |= thursday
		}
		if day == "Fri" {
			recurrence |= friday
		}
		if day == "Sat" {
			recurrence |= saturday
		}
		if day == "Sun" {
			recurrence |= sunday
		}
	}

	return fmt.Sprintf("W%03d/T%02s:%02s:00", recurrence, hours, minutes)
}
