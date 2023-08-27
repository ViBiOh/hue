package hue

import (
	"context"
	"fmt"
)

func (s *Service) listRules(ctx context.Context) (map[string]Rule, error) {
	var response map[string]Rule
	return response, get(ctx, fmt.Sprintf("%s/rules", s.bridgeURL), &response)
}

func (s *Service) createRule(ctx context.Context, o *Rule) error {
	id, err := create(ctx, fmt.Sprintf("%s/rules", s.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = id

	return nil
}

func (s *Service) deleteRule(ctx context.Context, id string) error {
	return remove(ctx, fmt.Sprintf("%s/rules/%s", s.bridgeURL, id))
}

func (s *Service) cleanRules(ctx context.Context) error {
	rules, err := s.listRules(ctx)
	if err != nil {
		return err
	}

	for key := range rules {
		if err := s.deleteRule(ctx, key); err != nil {
			return err
		}
	}

	return nil
}
