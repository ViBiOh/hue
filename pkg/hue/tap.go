package hue

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	v2 "github.com/ViBiOh/hue/pkg/v2"
)

var tapButtonMapping = map[string]string{
	"1": "34",
	"2": "16",
	"3": "17",
	"4": "18",
}

var rotaryTapButtonMapping = map[string]string{
	"1": "1000",
	"2": "2000",
	"3": "3000",
	"4": "4000",
}

func getButtonMapping(rotary bool, id string) string {
	if rotary {
		return rotaryTapButtonMapping[id]
	}
	return tapButtonMapping[id]
}

func (s *Service) createRuleDescription(groups []v2.Group, tapID string, rotary bool, button configTapButton) (Rule, error) {
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
				Value:    getButtonMapping(rotary, button.ID),
			},
		},
	}

	for _, group := range button.Groups {
		targetGroup, err := getGroup(groups, group)
		if err != nil {
			return Rule{}, err
		}

		newRule.Actions = append(newRule.Actions, Action{
			Address: fmt.Sprintf("/groups/%s/action", targetGroup.IDV1),
			Method:  http.MethodPut,
			Body:    States[button.State].V1(),
		})
	}

	for _, light := range button.Lights {
		newRule.Actions = append(newRule.Actions, Action{
			Address: fmt.Sprintf("/lights/%s/state", light),
			Method:  http.MethodPut,
			Body:    States[button.State].V1(),
		})
	}

	return newRule, nil
}

func (s *Service) configureTap(ctx context.Context, groups []v2.Group, taps []configTap) {
	for _, tap := range taps {
		for _, button := range tap.Buttons {
			rule, err := s.createRuleDescription(groups, tap.ID, tap.Rotary, button)
			if err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "create rule description", slog.Any("error", err))
			}

			if err := s.createRule(ctx, &rule); err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "create rule", slog.Any("error", err))
			}
		}
	}
}
