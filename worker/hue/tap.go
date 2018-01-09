package hue

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/iot/hue"
)

type tapConfig struct {
	ID      string
	Buttons []*tapButton
}

type tapButton struct {
	ID     string
	State  string
	Groups []string
	Rule   *rule
}

var (
	tapButtonMapping = map[string]string{
		`1`: `34`,
		`2`: `16`,
		`3`: `17`,
		`4`: `18`,
	}
)

func (a *App) createRuleDescription(button *tapButton) *rule {
	newRule := &rule{
		Name: fmt.Sprintf(`Tap %s.%s`, a.tap.ID, button.ID),
		Conditions: []*ruleCondition{
			&ruleCondition{
				Address:  fmt.Sprintf(`/sensors/%s/state/buttonevent`, a.tap.ID),
				Operator: `eq`,
				Value:    tapButtonMapping[button.ID],
			},
		},
		Actions: make([]*ruleAction, len(button.Groups)),
	}

	for index, group := range button.Groups {
		newRule.Actions[index] = &ruleAction{
			Address: fmt.Sprintf(`/groups/%s/action`, group),
			Method:  http.MethodPut,
			Body:    hue.States[button.State],
		}
	}

	return newRule
}

func (a *App) configureTap() {
	if err := a.cleanRules(); err != nil {
		log.Printf(`[hue] Error while cleaning rules: %v`, err)
	}

	for _, button := range a.tap.Buttons {
		button.Rule = a.createRuleDescription(button)
		if err := a.createRule(button.Rule); err != nil {
			log.Printf(`[hue] Error while creating rule: %v`, err)
			return
		}
	}
}
