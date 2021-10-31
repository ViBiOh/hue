package hue

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

var tapButtonMapping = map[string]string{
	"1": "34",
	"2": "16",
	"3": "17",
	"4": "18",
}

func (a *App) createRuleDescription(tapID string, button configTapButton) Rule {
	newRule := Rule{
		Name: fmt.Sprintf("Tap %s.%s", tapID, button.ID),
		Conditions: []Condition{
			{
				Address:  fmt.Sprintf("/sensors/%s/state/buttonevent", tapID),
				Operator: "dx",
			},
			{
				Address:  fmt.Sprintf("/sensors/%s/state/buttonevent", tapID),
				Operator: "eq",
				Value:    tapButtonMapping[button.ID],
			},
		},
		Actions: make([]Action, 0),
	}

	for _, group := range button.Groups {
		newRule.Actions = append(newRule.Actions, Action{
			Address: fmt.Sprintf("/groups/%s/action", group),
			Method:  http.MethodPut,
			Body:    States[button.State],
		})
	}

	return newRule
}

func (a *App) configureTap(ctx context.Context, taps []configTap) {
	for _, tap := range taps {
		for _, button := range tap.Buttons {
			button.Rule = a.createRuleDescription(tap.ID, button)
			if err := a.createRule(ctx, &button.Rule); err != nil {
				logger.Error("%s", err)
			}
		}
	}
}
