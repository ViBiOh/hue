package hue

import (
	"bytes"
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

type rule struct {
	ID         string
	Status     string
	Name       string
	Actions    []*ruleAction
	Conditions []*ruleCondition
}

type ruleAction struct {
	Address string
	Body    map[string]interface{}
	Method  string
}

type ruleCondition struct {
	Address  string
	Operator string
	Value    string
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
		return nil, fmt.Errorf(`Error while getting rules from bridge: %v`, err)
	}

	var rawRules map[string]*rule
	if err := json.Unmarshal(content, &rawRules); err != nil {
		return nil, fmt.Errorf(`Error while parsing rules from bridge: %v`, err)
	}

	var rules map[string]*rule
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

func (a *App) createRule(r *rule) error {
	content, err := httputils.RequestJSON(a.bridgeURL+`/rules`, r, nil, http.MethodPost)
	if err != nil || !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while creating rule: %s`, err)
	}

	var response []map[string]map[string]string
	if err := json.Unmarshal(content, &response); err != nil {
		return fmt.Errorf(`Error while unmarshalling create rule response: %s`, err)
	}

	r.ID = response[0][`success`][`id`]

	return nil
}

func (a *App) updateRule(r *rule) error {
	content, err := httputils.RequestJSON(a.bridgeURL+`/rules/`+r.ID, r, nil, http.MethodPut)
	if err != nil || !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while creating rule: %s`, err)
	}

	var response []map[string]map[string]string
	if err := json.Unmarshal(content, &response); err != nil {
		return fmt.Errorf(`Error while unmarshalling create rule response: %s`, err)
	}

	r.ID = response[0][`success`][`id`]

	return nil
}

func (a *App) createRuleDescription(button *tapButton, on bool) *rule {
	name := `On`
	if !on {
		name = `Off`
	}

	newRule := &rule{
		Status: `disabled`,
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
			Body: map[string]interface{}{
				`on`:             on,
				`transitiontime`: 30,
				`sat`:            0,
				`bri`:            254,
			},
		}
	}

	return newRule
}

func (a *App) deleteRule(id string) error {
	if _, err := httputils.Request(a.bridgeURL+`/rules+`+id, nil, nil, http.MethodDelete); err != nil {
		return fmt.Errorf(`Error while deleting rule from bridge: %v`, err)
	}

	return nil
}

func (a *App) cleanRules() error {
	rules, err := a.listRulesOfSensor(a.tap.ID)

	if err != nil {
		return fmt.Errorf(`Error while listing rules: %v`, err)
	}

	for key := range rules {
		if err := a.deleteRule(key); err != nil {
			return fmt.Errorf(`Error while deleting rule: %v`, err)
		}
	}

	return nil
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

		button.OnRule = on
		button.OffRule = off

		button.OnRule.Actions = append(button.OnRule.Actions, &ruleAction{
			Address: fmt.Sprintf(`/rules/%s`, button.OnRule.ID),
			Method:  http.MethodPut,
			Body: map[string]interface{}{
				`status`: `disabled`,
			},
		}, &ruleAction{
			Address: fmt.Sprintf(`/rules/%s`, button.OffRule.ID),
			Method:  http.MethodPut,
			Body: map[string]interface{}{
				`status`: `enabled`,
			},
		})

		button.OffRule.Actions = append(button.OffRule.Actions, &ruleAction{
			Address: fmt.Sprintf(`/rules/%s`, button.OffRule.ID),
			Method:  http.MethodPut,
			Body: map[string]interface{}{
				`status`: `disabled`,
			},
		}, &ruleAction{
			Address: fmt.Sprintf(`/rules/%s`, button.OnRule.ID),
			Method:  http.MethodPut,
			Body: map[string]interface{}{
				`status`: `enabled`,
			},
		})
	}
}
