package hue

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils"
)

type tapConfig struct {
	ID      string
	Buttons []*tapButton
}

type tapButton struct {
	ID      string
	Groups  []string
	OnRule  *rule
	OffRule *rule
}

var (
	tapButtonMapping = map[string]string{
		`1`: `34`,
		`2`: `16`,
		`3`: `17`,
		`4`: `18`,
	}
)

func (a *App) listRulesOfSensor(tapID string) (map[string]*rule, error) {
	content, err := httputils.GetRequest(a.bridgeURL+`/rules`, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting rules: %v`, err)
	}

	var rawRules map[string]*rule
	if err := json.Unmarshal(content, &rawRules); err != nil {
		return nil, fmt.Errorf(`Error while parsing rules: %v`, err)
	}

	rules := make(map[string]*rule)
	addressCondition := fmt.Sprintf(`/sensors/%s`, tapID)

	for id, r := range rawRules {
		match := false

		for _, condition := range r.Conditions {
			if strings.Contains(condition.Address, addressCondition) {
				match = true
				break
			}
		}

		if match {
			rules[id] = r
		}
	}

	return rules, nil
}

func (a *App) createRuleDescription(button *tapButton, on bool) *rule {
	name := `On`
	status := `enabled`
	body := map[string]interface{}{
		`on`:             true,
		`transitiontime`: 30,
		`sat`:            0,
		`bri`:            254,
	}

	if !on {
		name = `Off`
		status = `disabled`
		body = map[string]interface{}{
			`on`: false,
		}
	}

	newRule := &rule{
		Status: status,
		Name:   fmt.Sprintf(`Tap %s.%s - %s`, a.tap.ID, button.ID, name),
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
			Body:    body,
		}
	}

	return newRule
}

func (a *App) configureTap() {
	if err := a.cleanRules(); err != nil {
		log.Printf(`[hue] Error while cleaning rules: %v`, err)
	}

	for _, button := range a.tap.Buttons {
		on := a.createRuleDescription(button, true)
		if err := a.createRule(on); err != nil {
			log.Printf(`[hue] Error while creating on rule: %v`, err)
			return
		}

		off := a.createRuleDescription(button, false)
		if err := a.createRule(off); err != nil {
			log.Printf(`[hue] Error while creating off rule: %v`, err)
		}

		off.Actions = append(off.Actions, &ruleAction{
			Address: fmt.Sprintf(`/rules/%s`, off.ID),
			Method:  http.MethodPut,
			Body: map[string]interface{}{
				`status`: `disabled`,
			},
		}, &ruleAction{
			Address: fmt.Sprintf(`/rules/%s`, on.ID),
			Method:  http.MethodPut,
			Body: map[string]interface{}{
				`status`: `enabled`,
			},
		})
		if err := a.updateRule(off); err != nil {
			log.Printf(`[hue] Error while updating off rule: %v`, err)
		}

		on.Actions = append(on.Actions, &ruleAction{
			Address: fmt.Sprintf(`/rules/%s`, on.ID),
			Method:  http.MethodPut,
			Body: map[string]interface{}{
				`status`: `disabled`,
			},
		}, &ruleAction{
			Address: fmt.Sprintf(`/rules/%s`, off.ID),
			Method:  http.MethodPut,
			Body: map[string]interface{}{
				`status`: `enabled`,
			},
		})
		if err := a.updateRule(on); err != nil {
			log.Printf(`[hue] Error while updating on rule: %v`, err)
		}

		button.OnRule = on
		button.OffRule = off
	}
}
