package hue

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

const (
	presenceSensorType    = "ZLLPresence"
	temperatureSensorType = "ZLLTemperature"
)

func (a *app) listSensors(ctx context.Context) (map[string]Sensor, error) {
	var response map[string]Sensor

	if err := get(ctx, fmt.Sprintf("%s/sensors", a.bridgeURL), &response); err != nil {
		return nil, err
	}

	sensors := make(map[string]Sensor)

	for id, sensor := range response {
		if sensor.Type == presenceSensorType {
			sensor.ID = id
			sensors[sensor.Name] = sensor
		}
	}

	for _, sensor := range response {
		if sensor.Type == temperatureSensorType {
			if presenceSensor, ok := sensors[sensor.Name]; ok {
				presenceSensor.State.Temperature = sensor.State.Temperature / 100
				sensors[sensor.Name] = presenceSensor
			}
		}
	}

	return sensors, nil
}

func getGroupsActions(groups []string, state string) []Action {
	actions := make([]Action, 0)

	for _, group := range groups {
		actions = append(actions, Action{
			Address: fmt.Sprintf("/groups/%s/action", group),
			Method:  http.MethodPut,
			Body:    States[state],
		})
	}

	return actions
}

func (a *app) createSensorOnRuleDescription(sensor configSensor) Rule {
	state := "on"

	newRule := Rule{
		Name: fmt.Sprintf("MotionSensor %s - %s", sensor.ID, state),
		Conditions: []Condition{
			{
				Address:  fmt.Sprintf("/sensors/%s/state/presence", sensor.ID),
				Operator: "eq",
				Value:    "true",
			},
			{
				Address:  fmt.Sprintf("/sensors/%s/state/presence", sensor.ID),
				Operator: "dx",
			},
			{
				Address:  fmt.Sprintf("/sensors/%s/state/daylight", sensor.LightSensorID),
				Operator: "eq",
				Value:    "false",
			},
		},
		Actions: make([]Action, 0),
	}

	newRule.Actions = append(newRule.Actions, getGroupsActions(sensor.Groups, state)...)

	return newRule
}

func (a *app) createSensorOffRuleDescription(sensor configSensor) Rule {
	state := "long_off"

	newRule := Rule{
		Name: fmt.Sprintf("MotionSensor %s - %s", sensor.ID, state),
		Conditions: []Condition{
			{
				Address:  fmt.Sprintf("/sensors/%s/state/presence", sensor.ID),
				Operator: "eq",
				Value:    "false",
			},
			{
				Address:  fmt.Sprintf("/sensors/%s/state/presence", sensor.ID),
				Operator: "ddx",
				Value:    sensor.OffDelay,
			},
		},
		Actions: make([]Action, 0),
	}

	newRule.Actions = append(newRule.Actions, getGroupsActions(sensor.Groups, state)...)

	return newRule
}

func (a *app) configureMotionSensor(ctx context.Context, sensors []configSensor) {
	for _, sensor := range sensors {
		onRule := a.createSensorOnRuleDescription(sensor)
		if err := a.createRule(ctx, &onRule); err != nil {
			logger.Error("%s", err)
		}

		offRule := a.createSensorOffRuleDescription(sensor)
		if err := a.createRule(ctx, &offRule); err != nil {
			logger.Error("%s", err)
		}
	}
}

func (a *app) updateSensorConfig(ctx context.Context, sensor Sensor) error {
	if sensor.ID == "" {
		return errors.New("missing sensor ID to update")
	}

	return update(ctx, fmt.Sprintf("%s/sensors/%s/config", a.bridgeURL, sensor.ID), sensor.Config)
}
