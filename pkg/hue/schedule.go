package hue

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	v2 "github.com/ViBiOh/hue/pkg/v2"
)

func (s *Service) listSchedules(ctx context.Context) (map[string]Schedule, error) {
	var response map[string]Schedule

	if err := get(ctx, fmt.Sprintf("%s/schedules", s.bridgeURL), &response); err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	output := make(map[string]Schedule, len(response))
	for id, schedule := range response {
		schedule.ID = id
		output[id] = schedule
	}

	return output, nil
}

func (s *Service) createSchedule(ctx context.Context, o *Schedule) error {
	id, err := create(ctx, fmt.Sprintf("%s/schedules", s.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = id

	return nil
}

func (s *Service) createScheduleFromConfig(ctx context.Context, config ScheduleConfig, groups []v2.Group) error {
	targetGroup, err := getGroup(groups, config.Group)
	if err != nil {
		return err
	}

	scene, err := s.createSceneFromScheduleConfig(ctx, config, targetGroup)
	if err != nil {
		return err
	}

	schedule := &Schedule{
		APISchedule: APISchedule{
			Name:      config.Name,
			Localtime: config.Localtime,
			Command: Action{
				Address: fmt.Sprintf("/api/%s/groups/%s/action", s.bridgeUsername, targetGroup.IDV1),
				Body: map[string]any{
					"scene": scene.ID,
				},
				Method: http.MethodPut,
			},
		},
	}

	if err := s.createSchedule(ctx, schedule); err != nil {
		return err
	}

	return nil
}

func (s *Service) updateSchedule(ctx context.Context, schedule Schedule) error {
	if schedule.ID == "" {
		return errors.New("missing schedule ID to update")
	}

	return update(ctx, fmt.Sprintf("%s/schedules/%s", s.bridgeURL, schedule.ID), schedule.APISchedule)
}

func (s *Service) deleteSchedule(ctx context.Context, id string) error {
	return remove(ctx, fmt.Sprintf("%s/schedules/%s", s.bridgeURL, id))
}

func (s *Service) cleanSchedules(ctx context.Context) error {
	schedules, err := s.listSchedules(ctx)
	if err != nil {
		return err
	}

	for key := range schedules {
		if err := s.deleteSchedule(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) configureSchedules(ctx context.Context, groups []v2.Group, schedules []ScheduleConfig) {
	for _, config := range schedules {
		if err := s.createScheduleFromConfig(ctx, config, groups); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "create schedule", slog.Any("error", err))
		}
	}
}
