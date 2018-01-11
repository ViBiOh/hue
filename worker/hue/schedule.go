package hue

import (
	"fmt"
	"log"

	"github.com/ViBiOh/iot/hue"
)

type scheduleConfig struct {
	Name      string
	Localtime string
	Group     string
	State     string
}

func (a *App) listSchedules() (map[string]interface{}, error) {
	var response map[string]interface{}
	return response, get(fmt.Sprintf(`%s/schedules`, a.bridgeURL), response)
}

func (a *App) createSchedule(o *hue.Schedule) error {
	id, err := create(fmt.Sprintf(`%s/schedules`, a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *App) updateScheduleLightState(o *hue.Schedule, lightID string, state map[string]interface{}) error {
	return update(fmt.Sprintf(`%s/schedules/%s/lightstates/%s`, a.bridgeURL, o.ID, lightID), state)
}

func (a *App) deleteSchedule(id string) error {
	return delete(fmt.Sprintf(`%s/schedules/%s`, a.bridgeURL, id))
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
