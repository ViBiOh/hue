package hue

import (
	"context"
	"fmt"
)

func (a *app) listRules(ctx context.Context) (map[string]Rule, error) {
	var response map[string]Rule
	return response, get(ctx, fmt.Sprintf("%s/rules", a.bridgeURL), &response)
}

func (a *app) createRule(ctx context.Context, o *Rule) error {
	id, err := create(ctx, fmt.Sprintf("%s/rules", a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *app) updateRule(ctx context.Context, o Rule) error {
	return update(ctx, fmt.Sprintf("%s/rules/%s", a.bridgeURL, o.ID), o)
}

func (a *app) deleteRule(ctx context.Context, id string) error {
	return delete(ctx, fmt.Sprintf("%s/rules/%s", a.bridgeURL, id))
}

func (a *app) cleanRules(ctx context.Context) error {
	rules, err := a.listRules(ctx)
	if err != nil {
		return err
	}

	for key := range rules {
		if err := a.deleteRule(ctx, key); err != nil {
			return err
		}
	}

	return nil
}
