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

func (s *Service) createSensorOnRuleDescription(groups []v2.Group, motion v2.MotionSensor, sensor configSensor) (Rule, error) {
	state := "on"

	actions, err := getGroupsActions(groups, sensor, state)
	if err != nil {
		return Rule{}, fmt.Errorf("get groups actions: %w", err)
	}

	conditions := []Condition{
		{
			Address:  fmt.Sprintf(sensorPresenceURL, motion.IDV1),
			Operator: "eq",
			Value:    "true",
		},
		{
			Address:  fmt.Sprintf(sensorPresenceURL, motion.IDV1),
			Operator: "dx",
		},
	}

	if sensor.WhenDark {
		conditions = append(conditions, Condition{
			Address:  fmt.Sprintf("/sensors/%s/state/lightlevel", motion.LightLevelIDV1),
			Operator: "lt",
			Value:    "6000",
		})
	}

	if sensor.AllOff {
		for _, group := range sensor.Groups {
			targetGroup, err := getGroup(groups, group)
			if err != nil {
				return Rule{}, fmt.Errorf("get all groups off: %w", err)
			}

			conditions = append(conditions, Condition{
				Address:  fmt.Sprintf("/groups/%s/state/any_on", targetGroup.IDV1),
				Operator: "eq",
				Value:    "false",
			})
		}
	}

	newRule := Rule{
		Name:       fmt.Sprintf("MotionSensor %s - %s", motion.IDV1, state),
		Conditions: conditions,
		Actions:    actions,
	}

	return newRule, nil
}

func (s *Service) createSensorOffRuleDescription(groups []v2.Group, motion v2.MotionSensor, sensor configSensor) (Rule, error) {
	state := "long_off"

	actions, err := getGroupsActions(groups, sensor, state)
	if err != nil {
		return Rule{}, fmt.Errorf("get groups actions: %w", err)
	}

	newRule := Rule{
		Name: fmt.Sprintf("MotionSensor %s - %s", motion.IDV1, state),
		Conditions: []Condition{
			{
				Address:  fmt.Sprintf(sensorPresenceURL, motion.IDV1),
				Operator: "eq",
				Value:    "false",
			},
			{
				Address:  fmt.Sprintf(sensorPresenceURL, motion.IDV1),
				Operator: "ddx",
				Value:    sensor.OffDelay,
			},
		},
		Actions: actions,
	}

	return newRule, nil
}

func (s *Service) configureMotionSensor(ctx context.Context, groups []v2.Group, sensors []configSensor) {
	motionDevices := s.v2Service.Sensors()

	for _, sensor := range sensors {
		targetMotion, err := getMotionSensor(motionDevices, sensor.ID)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "unable to configure sensor", slog.String("id", sensor.ID), slog.Any("error", err))
			continue
		}

		onRule, err := s.createSensorOnRuleDescription(groups, targetMotion, sensor)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "create motion on rule", slog.Any("error", err))
		}

		if err := s.createRule(ctx, &onRule); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "create motion rule", slog.Any("error", err))
		}

		if len(sensor.OffDelay) != 0 {
			offRule, err := s.createSensorOffRuleDescription(groups, targetMotion, sensor)
			if err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "create motion off rule", slog.Any("error", err))
			}

			if err := s.createRule(ctx, &offRule); err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "create motion rule", slog.Any("error", err))
			}
		}
	}
}
