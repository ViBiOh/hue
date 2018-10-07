package hue

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/rollbar"
	"github.com/ViBiOh/iot/pkg/hue"
)

func (a *App) listSchedules(ctx context.Context) (map[string]*hue.Schedule, error) {
	var response map[string]*hue.Schedule

	if err := get(ctx, fmt.Sprintf(`%s/schedules`, a.bridgeURL), &response); err != nil {
		return response, nil
	}

	for id, schedule := range response {
		schedule.ID = id
	}

	return response, nil
}

func (a *App) getSchedule(ctx context.Context, id string) (*hue.Schedule, error) {
	var response hue.Schedule

	if err := get(ctx, fmt.Sprintf(`%s/schedules/%s`, a.bridgeURL, id), &response); err != nil {
		return &response, nil
	}

	return &response, nil
}

func (a *App) createSchedule(ctx context.Context, o *hue.Schedule) error {
	id, err := create(ctx, fmt.Sprintf(`%s/schedules`, a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *App) createScheduleFromConfig(ctx context.Context, config *hue.ScheduleConfig, groups map[string]*hue.Group) error {
	if groups == nil {
		var err error

		if groups, err = a.listGroups(ctx); err != nil {
			return fmt.Errorf(`error while retrieving groups for configuring schedule: %v`, err)
		}
	}

	scene, err := a.createSceneFromScheduleConfig(ctx, config, groups)
	if err != nil {
		return fmt.Errorf(`error while creating scene for schedule config %+v: %v`, config, err)
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

	if err := a.createSchedule(ctx, schedule); err != nil {
		return fmt.Errorf(`error while creating schedule from config %+v: %v`, config, err)
	}

	return nil
}

func (a *App) updateSchedule(ctx context.Context, schedule *hue.Schedule) error {
	if schedule == nil {
		return errors.New(`a schedule is required to update`)
	}

	if schedule.ID == `` {
		return errors.New(`a schedule ID is required to update`)
	}

	return update(ctx, fmt.Sprintf(`%s/schedules/%s`, a.bridgeURL, schedule.ID), schedule.APISchedule)
}

func (a *App) deleteSchedule(ctx context.Context, id string) error {
	return delete(ctx, fmt.Sprintf(`%s/schedules/%s`, a.bridgeURL, id))
}

func (a *App) cleanSchedules(ctx context.Context) error {
	schedules, err := a.listSchedules(ctx)
	if err != nil {
		return fmt.Errorf(`error while listing schedules: %v`, err)
	}

	for key := range schedules {
		if err := a.deleteSchedule(ctx, key); err != nil {
			return fmt.Errorf(`error while deleting schedule: %v`, err)
		}
	}

	return nil
}

func (a *App) configureSchedules(ctx context.Context, schedules []*hue.ScheduleConfig) {
	groups, err := a.listGroups(ctx)
	if err != nil {
		rollbar.LogError(`[%s] Error while retrieving groups for configuring schedules: %v`, hue.Source, err)
		return
	}

	for _, config := range schedules {
		if err := a.createScheduleFromConfig(ctx, config, groups); err != nil {
			rollbar.LogError(`[%s] %v`, hue.Source, err)
		}
	}
}