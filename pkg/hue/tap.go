package hue

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

var tapButtonMapping = map[string]string{
	"1": "34",
	"2": "16",
	"3": "17",
	"4": "18",
}

func (s *Service) createRuleDescription(tapID string, button configTapButton) Rule {
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
			Body:    States[button.State].V1(),
		})
	}

	return newRule
}

func (s *Service) configureTap(ctx context.Context, taps []configTap) {
	for _, tap := range taps {
		for _, button := range tap.Buttons {
			button.Rule = s.createRuleDescription(tap.ID, button)
			if err := s.createRule(ctx, &button.Rule); err != nil {
				slog.Error("create rule", "err", err)
			}
		}
	}
}
