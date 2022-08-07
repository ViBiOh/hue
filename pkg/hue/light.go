package hue

import (
	"context"
	"fmt"
)

func (a *App) listLights(ctx context.Context) (map[string]Light, error) {
	var response map[string]Light

	if err := get(ctx, fmt.Sprintf("%s/lights", a.bridgeURL), &response); err != nil {
		return nil, fmt.Errorf("get: %s", err)
	}

	output := make(map[string]Light, len(response))
	for id, light := range response {
		light.ID = id
		output[id] = light
	}

	return output, nil
}
