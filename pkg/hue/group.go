package hue

import (
	"context"
	"fmt"
	"strings"
)

func (a *App) listGroups(ctx context.Context) (map[string]Group, error) {
	var groups map[string]Group
	err := get(ctx, fmt.Sprintf("%s/groups", a.bridgeURL), &groups)
	if err != nil {
		return nil, err
	}

	output := make(map[string]Group, len(groups))

	for key, value := range groups {
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

		output[key] = value
	}

	return output, nil
}

func (a *App) updateGroupState(ctx context.Context, groupID string, state interface{}) error {
	return update(ctx, fmt.Sprintf("%s/groups/%s/action", a.bridgeURL, groupID), state)
}
