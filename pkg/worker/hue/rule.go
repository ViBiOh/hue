package hue

import (
	"context"
	"fmt"

	"github.com/ViBiOh/iot/pkg/hue"
)

func (a *App) listRules(ctx context.Context) (map[string]*hue.Rule, error) {
	var response map[string]*hue.Rule
	return response, get(ctx, fmt.Sprintf(`%s/rules`, a.bridgeURL), &response)
}

func (a *App) createRule(ctx context.Context, o *hue.Rule) error {
	id, err := create(ctx, fmt.Sprintf(`%s/rules`, a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *App) updateRule(ctx context.Context, o *hue.Rule) error {
	return update(ctx, fmt.Sprintf(`%s/rules/%s`, a.bridgeURL, o.ID), o)
}

func (a *App) deleteRule(ctx context.Context, id string) error {
	return delete(ctx, fmt.Sprintf(`%s/rules/%s`, a.bridgeURL, id))
}

func (a *App) cleanRules(ctx context.Context) error {
	rules, err := a.listRules(ctx)
	if err != nil {
		return fmt.Errorf(`error while listing rules: %v`, err)
	}

	for key := range rules {
		if err := a.deleteRule(ctx, key); err != nil {
			return fmt.Errorf(`error while deleting rule: %v`, err)
		}
	}

	return nil
}
