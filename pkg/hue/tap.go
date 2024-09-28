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

var dialTapButtonMapping = map[string]string{
	"1": "1000",
	"2": "2000",
	"3": "3000",
	"4": "4000",
}

var dialTapLongButtonMapping = map[string]string{
	"1": "1010",
	"2": "2010",
	"3": "3010",
	"4": "4010",
}

func getButtonMapping(dial bool, id string, long bool) string {
	if dial {
		if long {
			return dialTapLongButtonMapping[id]
		}
		return dialTapButtonMapping[id]
	}
	return tapButtonMapping[id]
}

func (s *Service) createRuleDescription(groups []v2.Group, tapID string, dial bool, button configTapButton) (Rule, error) {
	newRule := Rule{
		Name: fmt.Sprintf("Tap %s.%s.%t", tapID, button.ID, button.Long),
		Conditions: []Condition{
			{
				Address:  fmt.Sprintf("/sensors/%s/state/buttonevent", tapID),
				Operator: "dx",
			},
			{
				Address:  fmt.Sprintf("/sensors/%s/state/buttonevent", tapID),
				Operator: "eq",
				Value:    getButtonMapping(dial, button.ID, button.Long),
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
	tapDevices := s.v2Service.Taps()

	for _, tap := range taps {
		targetTap, err := getTap(tapDevices, tap.ID)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "unable to configure tap", slog.String("id", tap.ID), slog.Any("error", err))
			continue
		}

		for _, button := range tap.Buttons {
			rule, err := s.createRuleDescription(groups, targetTap.IDV1, targetTap.Dial, button)
			if err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "create tap rule description", slog.Any("error", err))
			}

			if err := s.createRule(ctx, &rule); err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "create tap rule", slog.Any("error", err))
			}
		}
	}
}
