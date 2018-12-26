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
		},
		Actions: make([]*hue.Action, 0),
	}

	for _, group := range sensor.Groups {
		newRule.Actions = append(newRule.Actions, &hue.Action{
			Address: fmt.Sprintf(`/groups/%s/action`, group),
			Method:  http.MethodPut,
			Body:    hue.States[state],
		})
	}

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
				Address:  fmt.Sprintf(`/sensors/%s/state/presence`, sensor.ID),
				Operator: `ddx`,
				Value:    sensor.OffDelay,
			},
		},
		Actions: make([]*hue.Action, 0),
	}

	for _, group := range sensor.Groups {
		newRule.Actions = append(newRule.Actions, &hue.Action{
			Address: fmt.Sprintf(`/groups/%s/action`, group),
			Method:  http.MethodPut,
			Body:    hue.States[state],
		})
	}

	return newRule
}

func (a *App) configureMotionSensor(ctx context.Context, sensors []*sensorConfig) {
	for _, sensor := range sensors {
		onRule := a.createSensorOnRuleDescription(sensor)
		if err := a.createRule(ctx, onRule); err != nil {
			logger.Error(`%+v`, err)
		}

		offRule := a.createSensorOffRuleDescription(sensor)
		if err := a.createRule(ctx, offRule); err != nil {
			logger.Error(`%+v`, err)
		}
	}
}
