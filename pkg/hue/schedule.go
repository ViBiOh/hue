package hue

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func (a *app) listSchedules(ctx context.Context) (map[string]Schedule, error) {
	var response map[string]Schedule

	if err := get(ctx, fmt.Sprintf("%s/schedules", a.bridgeURL), &response); err != nil {
		return nil, err
	}

	output := make(map[string]Schedule, len(response))
	for id, schedule := range response {
		schedule.ID = id
		output[id] = schedule
	}

	return output, nil
}

func (a *app) createSchedule(ctx context.Context, o *Schedule) error {
	id, err := create(ctx, fmt.Sprintf("%s/schedules", a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = id

	return nil
}

func (a *app) createScheduleFromConfig(ctx context.Context, config ScheduleConfig, groups map[string]Group) error {
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

	schedule := &Schedule{
		APISchedule: APISchedule{
			Name:      config.Name,
			Localtime: config.Localtime,
			Command: Action{
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

func (a *app) updateSchedule(ctx context.Context, schedule Schedule) error {
	if schedule.ID == "" {
		return errors.New("missing schedule ID to update")
	}

	return update(ctx, fmt.Sprintf("%s/schedules/%s", a.bridgeURL, schedule.ID), schedule.APISchedule)
}

func (a *app) deleteSchedule(ctx context.Context, id string) error {
	return remove(ctx, fmt.Sprintf("%s/schedules/%s", a.bridgeURL, id))
}

func (a *app) cleanSchedules(ctx context.Context) error {
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

func (a *app) configureSchedules(ctx context.Context, schedules []ScheduleConfig) {
	groups, err := a.listGroups(ctx)
	if err != nil {
		logger.Error("%s", err)
		return
	}

	for _, config := range schedules {
		if err := a.createScheduleFromConfig(ctx, config, groups); err != nil {
			logger.Error("%s", err)
		}
	}
}
