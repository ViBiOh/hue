package hue

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/iot/pkg/hue"
)

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
				Address:  fmt.Sprintf(`/sensors/%s/state/status`, sensor.CompanionID),
				Operator: `eq`,
				Value:    `0`,
			},
		},
		Actions: make([]*hue.Action, len(sensor.Groups)+1),
	}

	for index, group := range sensor.Groups {
		newRule.Actions[index] = &hue.Action{
			Address: fmt.Sprintf(`/groups/%s/action`, group),
			Method:  http.MethodPut,
			Body:    hue.States[state],
		}
	}

	newRule.Actions[len(sensor.Groups)] = &hue.Action{
		Address: fmt.Sprintf(`/sensors/%s/state`, sensor.CompanionID),
		Method:  http.MethodPut,
		Body: map[string]interface{}{
			`status`: 1,
		},
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
			{
				Address:  fmt.Sprintf(`/sensors/%s/state/status`, sensor.CompanionID),
				Operator: `eq`,
				Value:    `1`,
			},
		},
		Actions: make([]*hue.Action, len(sensor.Groups)+1),
	}

	for index, group := range sensor.Groups {
		newRule.Actions[index] = &hue.Action{
			Address: fmt.Sprintf(`/groups/%s/action`, group),
			Method:  http.MethodPut,
			Body:    hue.States[state],
		}
	}

	newRule.Actions[len(sensor.Groups)] = &hue.Action{
		Address: fmt.Sprintf(`/sensors/%s/state`, sensor.CompanionID),
		Method:  http.MethodPut,
		Body: map[string]interface{}{
			`status`: 0,
		},
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
