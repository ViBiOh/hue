package hue

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	v2 "github.com/ViBiOh/hue/pkg/v2"
)

const sensorPresenceURL = "/sensors/%s/state/presence"

func getGroupsActions(groups []v2.Group, config configSensor, state string) ([]Action, error) {
	var actions []Action

	for _, group := range config.Groups {
		targetGroup, err := getGroup(groups, group)
		if err != nil {
			return nil, err
		}

		actions = append(actions, Action{
			Address: fmt.Sprintf("/groups/%s/action", targetGroup.IDV1),
			Method:  http.MethodPut,
			Body:    States[state].V1(),
		})
	}

	return actions, nil
}

func (s *Service) createSensorOnRuleDescription(sensor configSensor, groups []v2.Group) (Rule, error) {
	state := "on"

	actions, err := getGroupsActions(groups, sensor, state)
	if err != nil {
		return Rule{}, fmt.Errorf("get groups actions: %w", err)
	}

	newRule := Rule{
		Name: fmt.Sprintf("MotionSensor %s - %s", sensor.ID, state),
		Conditions: []Condition{
			{
				Address:  fmt.Sprintf(sensorPresenceURL, sensor.ID),
				Operator: "eq",
				Value:    "true",
			},
			{
				Address:  fmt.Sprintf(sensorPresenceURL, sensor.ID),
				Operator: "dx",
			},
			{
				Address:  fmt.Sprintf("/sensors/%s/state/daylight", sensor.LightSensorID),
				Operator: "eq",
				Value:    "false",
			},
		},
		Actions: actions,
	}

	return newRule, nil
}

func (s *Service) createSensorOffRuleDescription(sensor configSensor, groups []v2.Group) (Rule, error) {
	state := "long_off"

	actions, err := getGroupsActions(groups, sensor, state)
	if err != nil {
		return Rule{}, fmt.Errorf("get groups actions: %w", err)
	}

	newRule := Rule{
		Name: fmt.Sprintf("MotionSensor %s - %s", sensor.ID, state),
		Conditions: []Condition{
			{
				Address:  fmt.Sprintf(sensorPresenceURL, sensor.ID),
				Operator: "eq",
				Value:    "false",
			},
			{
				Address:  fmt.Sprintf(sensorPresenceURL, sensor.ID),
				Operator: "ddx",
				Value:    sensor.OffDelay,
			},
		},
		Actions: actions,
	}

	return newRule, nil
}

func (s *Service) configureMotionSensor(ctx context.Context, sensors []configSensor) {
	groups := s.v2Service.Groups()

	for _, sensor := range sensors {
		onRule, err := s.createSensorOnRuleDescription(sensor, groups)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "create sensor on rule", slog.Any("error", err))
		}

		if err := s.createRule(ctx, &onRule); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "create rule", slog.Any("error", err))
		}

		offRule, err := s.createSensorOffRuleDescription(sensor, groups)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "create sensor off rule", slog.Any("error", err))
		}

		if err := s.createRule(ctx, &offRule); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "create rule", slog.Any("error", err))
		}
	}
}
