package hue

import (
	"fmt"

	"github.com/ViBiOh/iot/hue"
)

func (a *App) listRules() (map[string]*hue.Rule, error) {
	var response map[string]*hue.Rule
	return response, get(fmt.Sprintf(`%s/rules`, a.bridgeURL), &response)
}

func (a *App) createRule(o *hue.Rule) error {
	id, err := create(fmt.Sprintf(`%s/rules`, a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *App) updateRule(o *hue.Rule) error {
	return update(fmt.Sprintf(`%s/rules/%s`, a.bridgeURL, o.ID), o)
}

func (a *App) deleteRule(id string) error {
	return delete(fmt.Sprintf(`%s/rules/%s`, a.bridgeURL, id))
}

func (a *App) cleanRules() error {
	rules, err := a.listRules()
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
