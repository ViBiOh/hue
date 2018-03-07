package hue

import (
	"errors"
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

func (a *App) getSchedule(id string) (*hue.Schedule, error) {
	var response hue.Schedule

	if err := get(fmt.Sprintf(`%s/schedules/%s`, a.bridgeURL, id), &response); err != nil {
		return &response, nil
	}

	return &response, nil
}

func (a *App) createSchedule(o *hue.Schedule) error {
	id, err := create(fmt.Sprintf(`%s/schedules`, a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *App) createScheduleFromConfig(config *hue.ScheduleConfig, groups map[string]*hue.Group) error {
	if groups == nil {
		var err error

		if groups, err = a.listGroups(); err != nil {
			return fmt.Errorf(`Error while retrieving groups for configuring schedule: %v`, err)
		}
	}

	scene, err := a.createSceneFromScheduleConfig(config, groups)
	if err != nil {
		return fmt.Errorf(`Error while creating scene for schedule config %+v: %v`, config, err)
	}

	schedule := &hue.Schedule{
		APISchedule: &hue.APISchedule{
			Name:      config.Name,
			Localtime: config.Localtime,
			Command: &hue.Action{
				Address: fmt.Sprintf(`/api/%s/groups/%s/action`, a.bridgeUsername, config.Group),
				Body: map[string]interface{}{
					`scene`: scene.ID,
				},
				Method: http.MethodPut,
			},
		},
	}

	if err := a.createSchedule(schedule); err != nil {
		return fmt.Errorf(`Error while creating schedule from config %+v: %v`, config, err)
	}

	return nil
}

func (a *App) updateSchedule(schedule *hue.Schedule) error {
	if schedule == nil {
		return errors.New(`A schedule is required to update`)
	}

	if schedule.ID == `` {
		return errors.New(`A schedule ID is required to update`)
	}

	return update(fmt.Sprintf(`%s/schedules/%s`, a.bridgeURL, schedule.ID), schedule.APISchedule)
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

func (a *App) configureSchedules(schedules []*hue.ScheduleConfig) {
	groups, err := a.listGroups()
	if err != nil {
		log.Printf(`[%s] Error while retrieving groups for configuring schedules: %v`, hue.HueSource, err)
		return
	}

	for _, config := range schedules {
		if err := a.createScheduleFromConfig(config, groups); err != nil {
			log.Printf(`[%s] %v`, hue.HueSource, err)
		}
	}
}
