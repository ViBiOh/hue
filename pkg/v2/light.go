package v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

type Light struct {
	ID       string `json:"id"`
	Metadata struct {
		Archetype string `json:"archetype"`
		Name      string `json:"name"`
	} `json:"metadata"`
	On               On               `json:"on"`
	Dimming          Dimming          `json:"dimming"`
	Color            Color            `json:"color"`
	ColorTemperature ColorTemperature `json:"color_temperature"`
}

type Dimming struct {
	Brightness float64 `json:"brightness"`
}

type On struct {
	On bool `json:"on"`
}

func (a *App) buildLights(ctx context.Context) (map[string]*Light, error) {
	lights, err := list[Light](ctx, a.req, "light")
	if err != nil {
		return nil, fmt.Errorf("list lights: %w", err)
	}

	output := make(map[string]*Light, len(lights))
	for _, light := range lights {
		light := light
		output[light.ID] = &light

		if err := a.setWhiteLight(ctx, light.ID); err != nil {
			logger.Error("white light: %s", err)
		}
	}

	return output, err
}

func (a *App) setWhiteLight(ctx context.Context, id string) error {
	var color Color
	color.XY.X = 0.37203
	color.XY.Y = 0.37763

	var colorTemperature ColorTemperature
	colorTemperature.Mirek = 238 // 4200K

	payload := map[string]interface{}{
		"color":             color,
		"color_temperature": colorTemperature,
	}

	if _, err := a.req.Method(http.MethodPut).Path("/clip/v2/resource/light/"+id).JSON(ctx, payload); err != nil {
		return fmt.Errorf("update light `%s`: %w", id, err)
	}

	return nil
}
