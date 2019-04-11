package hue

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/iot/pkg/hue"
)

func (a *App) listSchedules(ctx context.Context) (map[string]*hue.Schedule, error) {
	var response map[string]*hue.Schedule

	if err := get(ctx, fmt.Sprintf("%s/schedules", a.bridgeURL), &response); err != nil {
		return nil, err
	}

	for id, schedule := range response {
		schedule.ID = id
	}

	return response, nil
}

func (a *App) getSchedule(ctx context.Context, id string) (*hue.Schedule, error) {
	var response hue.Schedule

	if err := get(ctx, fmt.Sprintf("%s/schedules/%s", a.bridgeURL, id), &response); err != nil {
		return &response, nil
	}

	return &response, nil
}

func (a *App) createSchedule(ctx context.Context, o *hue.Schedule) error {
	id, err := create(ctx, fmt.Sprintf("%s/schedules", a.bridgeURL), o)
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
			return err
		}
	}

	scene, err := a.createSceneFromScheduleConfig(ctx, config, groups)
	if err != nil {
		return err
	}

	schedule := &hue.Schedule{
		APISchedule: &hue.APISchedule{
			Name:      config.Name,
			Localtime: config.Localtime,
			Command: &hue.Action{
				Address: fmt.Sprintf("/api/%s/groups/%s/action", a.bridgeUsername, config.Group),
				Body: map[string]interface{}{
					"scene": scene.ID,
				},
				Method: http.MethodPut,
			},
		},
	}

	if err := a.createSchedule(ctx, schedule); err != nil {
		return err
	}

	return nil
}

func (a *App) updateSchedule(ctx context.Context, schedule *hue.Schedule) error {
	if schedule == nil {
		return errors.New("missing schedule to update")
	}

	if schedule.ID == "" {
		return errors.New("missing schedule ID to update")
	}

	return update(ctx, fmt.Sprintf("%s/schedules/%s", a.bridgeURL, schedule.ID), schedule.APISchedule)
}

func (a *App) deleteSchedule(ctx context.Context, id string) error {
	return delete(ctx, fmt.Sprintf("%s/schedules/%s", a.bridgeURL, id))
}

func (a *App) cleanSchedules(ctx context.Context) error {
	schedules, err := a.listSchedules(ctx)
	if err != nil {
		return err
	}

	for key := range schedules {
		if err := a.deleteSchedule(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) configureSchedules(ctx context.Context, schedules []*hue.ScheduleConfig) {
	groups, err := a.listGroups(ctx)
	if err != nil {
		logger.Error("%+v", err)
		return
	}

	for _, config := range schedules {
		if err := a.createScheduleFromConfig(ctx, config, groups); err != nil {
			logger.Error("%+v", err)
		}
	}
}
