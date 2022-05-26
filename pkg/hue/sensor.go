package hue

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

const sensorPresenceURL = "/sensors/%s/state/presence"

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

func (a *App) createSensorOnRuleDescription(sensor configSensor) Rule {
	state := "on"

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
		Actions: make([]Action, 0),
	}

	newRule.Actions = append(newRule.Actions, getGroupsActions(sensor.Groups, state)...)

	return newRule
}

func (a *App) createSensorOffRuleDescription(sensor configSensor) Rule {
	state := "long_off"

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
		Actions: make([]Action, 0),
	}

	newRule.Actions = append(newRule.Actions, getGroupsActions(sensor.Groups, state)...)

	return newRule
}

func (a *App) configureMotionSensor(ctx context.Context, sensors []configSensor) {
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
