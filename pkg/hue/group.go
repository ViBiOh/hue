package hue

import (
	"context"
	"fmt"
	"strings"
)

func (a *App) listGroups(ctx context.Context) (map[string]Group, error) {
	var response map[string]Group
	err := get(ctx, fmt.Sprintf("%s/groups", a.bridgeURL), &response)
	if err != nil {
		return nil, fmt.Errorf("unable to get: %s", err)
	}

	output := make(map[string]Group, len(response))

	for key, value := range response {
		value.Tap = false

		for _, lightID := range value.Lights {
			if strings.HasPrefix(a.lights[lightID].Type, "On/Off") {
				value.Tap = true
			}
		}

		value.ID = key
		output[key] = value
	}

	return output, nil
}
