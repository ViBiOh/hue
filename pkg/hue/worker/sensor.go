package hue

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/iot/pkg/hue"
)

const (
	presenceSensorType    = `ZLLPresence`
	temperatureSensorType = `ZLLTemperature`
	dimmedDelay           = `PT00:00:15`
)

func (a *App) listSensors(ctx context.Context) (map[string]*hue.Sensor, error) {
	var response map[string]*hue.Sensor

	if err := get(ctx, fmt.Sprintf(`%s/sensors`, a.bridgeURL), &response); err != nil {
		return nil, err
	}

	sensors := make(map[string]*hue.Sensor)

	for _, sensor := range response {
		if sensor.Type == presenceSensorType {
			sensors[sensor.Name] = sensor
		}
	}

	for _, sensor := range response {
		if sensor.Type == temperatureSensorType {
			if presenceSensor, ok := sensors[sensor.Name]; ok {
				presenceSensor.State.Temperature = sensor.State.Temperature / 100
			}
		}
	}

	return sensors, nil
}

func getStatusAction(id string, status int) *hue.Action {
	return &hue.Action{
		Address: fmt.Sprintf(`/sensors/%s/state`, id),
		Method:  http.MethodPut,
		Body: map[string]interface{}{
			`status`: status,
		},
	}
}

func getGroupsActions(groups []string, state string) []*hue.Action {
	actions := make([]*hue.Action, 0)

	for _, group := range groups {
		actions = append(actions, &hue.Action{
			Address: fmt.Sprintf(`/groups/%s/action`, group),
			Method:  http.MethodPut,
			Body:    hue.States[state],
		})
	}

	return actions
}

func (a *App) createSensorOnRuleDescription(sensor *sensorConfig) *hue.Rule {
	state := `on`

	newRule := &hue.Rule{
		Name: fmt.Sprintf(`MotionSensor %s - %s`, sensor.ID, state),
		Conditions: []*hue.Condition{
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/presence`, sensor.ID),
				Operator: `eq`,
				Value:    `true`,
			},
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/presence`, sensor.ID),
				Operator: `dx`,
			},
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/dark`, sensor.LightSensorID),
				Operator: `eq`,
				Value:    `true`,
			},
		},
		Actions: make([]*hue.Action, 0),
	}

	newRule.Actions = append(newRule.Actions, getStatusAction(sensor.CompanionID, 1))
	newRule.Actions = append(newRule.Actions, getGroupsActions(sensor.Groups, state)...)

	return newRule
}

func (a *App) createSensorRecoverRuleDescription(sensor *sensorConfig) *hue.Rule {
	newRule := &hue.Rule{
		Name: fmt.Sprintf(`MotionSensor %s - recover`, sensor.ID),
		Conditions: []*hue.Condition{
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/presence`, sensor.ID),
				Operator: `eq`,
				Value:    `true`,
			},
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/presence`, sensor.ID),
				Operator: `dx`,
			},
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/status`, sensor.CompanionID),
				Operator: `gt`,
				Value:    `0`,
			},
		},
		Actions: make([]*hue.Action, 0),
	}

	newRule.Actions = append(newRule.Actions, getStatusAction(sensor.CompanionID, 1))
	newRule.Actions = append(newRule.Actions, getGroupsActions(sensor.Groups, `on`)...)

	return newRule
}

func (a *App) createSensorDimmedRuleDescription(sensor *sensorConfig) *hue.Rule {
	state := `dimmed`

	newRule := &hue.Rule{
		Name: fmt.Sprintf(`MotionSensor %s - %s`, sensor.ID, state),
		Conditions: []*hue.Condition{
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/presence`, sensor.ID),
				Operator: `eq`,
				Value:    `false`,
			},
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/presence`, sensor.ID),
				Operator: `ddx`,
				Value:    sensor.OffDelay,
			},
		},
		Actions: make([]*hue.Action, 0),
	}

	newRule.Actions = append(newRule.Actions, getStatusAction(sensor.CompanionID, 2))
	newRule.Actions = append(newRule.Actions, getGroupsActions(sensor.Groups, state)...)

	return newRule
}

func (a *App) createSensorOffRuleDescription(sensor *sensorConfig) *hue.Rule {
	state := `off`

	newRule := &hue.Rule{
		Name: fmt.Sprintf(`MotionSensor %s - %s`, sensor.ID, state),
		Conditions: []*hue.Condition{
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/presence`, sensor.ID),
				Operator: `eq`,
				Value:    `false`,
			},
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/status`, sensor.CompanionID),
				Operator: `ddx`,
				Value:    dimmedDelay,
			},
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/status`, sensor.CompanionID),
				Operator: `gt`,
				Value:    `1`,
			},
		},
		Actions: make([]*hue.Action, 0),
	}

	newRule.Actions = append(newRule.Actions, getStatusAction(sensor.CompanionID, 0))
	newRule.Actions = append(newRule.Actions, getGroupsActions(sensor.Groups, state)...)

	return newRule
}

func (a *App) configureMotionSensor(ctx context.Context, sensors []*sensorConfig) {
	for _, sensor := range sensors {
		onRule := a.createSensorOnRuleDescription(sensor)
		if err := a.createRule(ctx, onRule); err != nil {
			logger.Error(`%+v`, err)
		}

		recoverRule := a.createSensorRecoverRuleDescription(sensor)
		if err := a.createRule(ctx, recoverRule); err != nil {
			logger.Error(`%+v`, err)
		}

		dimmedRule := a.createSensorDimmedRuleDescription(sensor)
		if err := a.createRule(ctx, dimmedRule); err != nil {
			logger.Error(`%+v`, err)
		}

		offRule := a.createSensorOffRuleDescription(sensor)
		if err := a.createRule(ctx, offRule); err != nil {
			logger.Error(`%+v`, err)
		}
	}
}
