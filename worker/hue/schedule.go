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
		return nil, fmt.Errorf(`Error while sending get request: %v`, err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(content, &response); err != nil {
		return nil, fmt.Errorf(`Error while parsing response: %v`, err)
	}

	return response, nil
}

func (a *App) createSchedule(o *hue.Schedule) error {
	content, err := httputils.RequestJSON(fmt.Sprintf(`%s/schedules`, a.bridgeURL), o, nil, http.MethodPost)
	if err != nil {
		return fmt.Errorf(`Error while sending post request: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while sending post request: %s`, content)
	}

	var response []map[string]map[string]string
	if err := json.Unmarshal(content, &response); err != nil {
		return fmt.Errorf(`Error while parsing response: %s`, err)
	}

	o.ID = response[0][`success`][`id`]

	return nil
}

func (a *App) updateScheduleLightState(o *hue.Schedule, lightID string, state map[string]interface{}) error {
	content, err := httputils.RequestJSON(fmt.Sprintf(`%s/schedules/%s/lightstates/%s`, a.bridgeURL, o.ID, lightID), state, nil, http.MethodPut)
	if err != nil {
		return fmt.Errorf(`Error while sending put request: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while sending put request: %s`, content)
	}

	return nil
}

func (a *App) deleteSchedule(id string) error {
	content, err := httputils.Request(fmt.Sprintf(`%s/schedules/%s`, a.bridgeURL, id), nil, nil, http.MethodDelete)
	if err != nil {
		return fmt.Errorf(`Error while sending delete request: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while sending delete request: %s`, content)
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
		log.Printf(`[hue] Error while retrieving groups for configuring schedules: %v`, err)
		return
	}

	for _, config := range schedules {
		group, ok := groups[config.Group]
		if !ok {
			log.Printf(`[hue] Unknown group id: %s`, config.Group)
			continue
		}

		state, ok := hue.States[config.State]
		if !ok {
			log.Printf(`[hue] Unknown state name: %s`, config.State)
			continue
		}

		schedule := &hue.Schedule{
			Name:      config.Name,
			Localtime: config.Localtime,
			Lights:    group.Lights,
		}

		if err := a.createSchedule(schedule); err != nil {
			log.Printf(`[hue] Error while creating schedule: %v`, err)
		}

		for _, light := range schedule.Lights {
			a.updateScheduleLightState(schedule, light, state)
		}
	}
}
