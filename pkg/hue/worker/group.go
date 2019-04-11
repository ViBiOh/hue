package hue

import (
	"context"
	"fmt"
	"strings"

	"github.com/ViBiOh/iot/pkg/hue"
)

func (a *App) listGroups(ctx context.Context) (map[string]*hue.Group, error) {
	var groups map[string]*hue.Group
	err := get(ctx, fmt.Sprintf("%s/groups", a.bridgeURL), &groups)
	if err != nil {
		return nil, err
	}

	for _, value := range groups {
		value.Tap = false

		for _, lightID := range value.Lights {
			light, err := a.getLight(ctx, lightID)
			if err != nil {
				return nil, err
			}

			if strings.HasPrefix(light.Type, "On/Off") {
				value.Tap = true
			}
		}
	}

	return groups, nil
}

func (a *App) updateGroupState(ctx context.Context, groupID string, state interface{}) error {
	return update(ctx, fmt.Sprintf("%s/groups/%s/action", a.bridgeURL, groupID), state)
}
