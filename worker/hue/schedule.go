package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/iot/hue"
)

type scheduleConfig struct {
	Name      string
	Localtime string
	Group     string
	State     string
}

func (a *App) listSchedules() (map[string]interface{}, error) {
	content, err := httputils.GetRequest(fmt.Sprintf(`%s/schedules`, a.bridgeURL), nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting schedules: %v`, err)
	}

	var schedules map[string]interface{}
	if err := json.Unmarshal(content, &schedules); err != nil {
		return nil, fmt.Errorf(`Error while parsing schedules: %v`, err)
	}

	return schedules, nil
}

func (a *App) createSchedule(s *hue.Schedule) error {
	content, err := httputils.RequestJSON(fmt.Sprintf(`%s/schedules`, a.bridgeURL), s, nil, http.MethodPost)
	if err != nil {
		return fmt.Errorf(`Error while send schedule creation: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error during schedule creation: %s`, content)
	}

	var response []map[string]map[string]string
	if err := json.Unmarshal(content, &response); err != nil {
		return fmt.Errorf(`Error while unmarshalling create schedule response: %s`, err)
	}

	s.ID = response[0][`success`][`id`]

	return nil
}

func (a *App) updateScheduleLightState(s *hue.Schedule, lightID string, state map[string]interface{}) error {
	content, err := httputils.RequestJSON(fmt.Sprintf(`%s/schedules/%s/lightstates/%s`, a.bridgeURL, s.ID, lightID), state, nil, http.MethodPost)
	if err != nil {
		return fmt.Errorf(`Error while updating light state of schedule: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while updating light state of schedule: %s`, content)
	}

	return nil
}

func (a *App) deleteSchedule(id string) error {
	content, err := httputils.Request(a.bridgeURL+`/schedules/`+id, nil, nil, http.MethodDelete)
	if err != nil {
		return fmt.Errorf(`Error while deleting schedule: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while deleting schedule: %s`, content)
	}

	return nil
}

func (a *App) cleanSchedules() error {
	schedules, err := a.listSchedules()
	if err != nil {
		return fmt.Errorf(`Error while listing schedules: %v`, err)
	}

	for key := range schedules {
		if err := a.deleteSchedule(key); err != nil {
			return fmt.Errorf(`Error while deleting schedule: %v`, err)
		}
	}

	return nil
}

func (a *App) configureSchedule(schedules []*scheduleConfig) {
	groups, err := a.getGroups()
	if err != nil {
		log.Printf(`[hue] Error while retrieving groups for configuring schedule: %v`, err)
	}

	for _, config := range schedules {
		if group, ok := groups[config.Group]; ok {
			schedule := &hue.Schedule{
				Name:      config.Name,
				Localtime: config.Localtime,
				Lights:    group.Lights,
			}

			if err := a.createSchedule(schedule); err != nil {
				log.Printf(`[hue] Error while creating schedule: %v`, err)
			}

			if state, ok := hue.States[config.State]; ok {
				for _, light := range schedule.Lights {
					a.updateScheduleLightState(schedule, light, state)
				}
			} else {
				log.Printf(`[hue] Unknown state name: %s`, config.State)
			}
		} else {
			log.Printf(`[hue] Unknown group id: %s`, config.Group)
		}
	}
}
