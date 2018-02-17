package hue

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/iot/hue"
)

func (a *App) listSchedules() (map[string]*hue.Schedule, error) {
	var response map[string]*hue.Schedule

	if err := get(fmt.Sprintf(`%s/schedules`, a.bridgeURL), &response); err != nil {
		return response, nil
	}

	for id, schedule := range response {
		schedule.ID = id
	}

	return response, nil
}

func (a *App) createSchedule(o *hue.Schedule) error {
	id, err := create(fmt.Sprintf(`%s/schedules`, a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *App) updateScheduleStatus(id, status string) error {
	return update(fmt.Sprintf(`%s/schedules/%s`, a.bridgeURL, id), map[string]string{
		`status`: status,
	})
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
	groups, err := a.listGroups()
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

		scene := &hue.Scene{
			Name:    config.Name,
			Lights:  group.Lights,
			Recycle: false,
		}

		if err := a.createScene(scene); err != nil {
			log.Printf(`[hue] Error while creating scene: %v`, err)
			continue
		}

		for _, light := range scene.Lights {
			if err := a.updateSceneLightState(scene, light, state); err != nil {
				log.Printf(`[hue] Error while updating scene light state: %v`, err)
			}
		}

		schedule := &hue.Schedule{
			Name:      config.Name,
			Localtime: config.Localtime,
			Command: &hue.Action{
				Address: fmt.Sprintf(`/api/%s/groups/%s/action`, a.bridgeUsername, config.Group),
				Body: map[string]interface{}{
					`scene`: scene.ID,
				},
				Method: http.MethodPut,
			},
		}

		if err := a.createSchedule(schedule); err != nil {
			log.Printf(`[hue] Error while creating schedule: %v`, err)
			continue
		}
	}
}
