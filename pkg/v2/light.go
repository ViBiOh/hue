package v2

import (
	"context"
	"fmt"
)

// Light description
type Light struct {
	ID       string `json:"id"`
	Metadata struct {
		Archetype string `json:"archetype"`
		Name      string `json:"name"`
	} `json:"metadata"`
	On      On      `json:"on"`
	Dimming Dimming `json:"dimming"`
}

// Dimming description
type Dimming struct {
	Brightness float64 `json:"brightness"`
}

// On description
type On struct {
	On bool `json:"on"`
}

func (a *App) buildLights(ctx context.Context) (map[string]*Light, error) {
	lights, err := list[Light](ctx, a.req, "light")
	if err != nil {
		return nil, fmt.Errorf("unable to list lights: %s", err)
	}

	output := make(map[string]*Light, len(lights))
	for _, light := range lights {
		light := light
		output[light.ID] = &light
	}

	return output, err
}
