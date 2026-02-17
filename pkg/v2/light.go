package v2

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"
)

type Light struct {
	ID       string `json:"id"`
	IDV1     string `json:"id_v1"`
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

var temperatures = map[string]int{
	"warm":    int(math.Round(1000000 / 2700)),
	"soft":    int(math.Round(1000000 / 3000)),
	"neutral": int(math.Round(1000000 / 4000)),
	"cool":    int(math.Round(1000000 / 5000)),
}

var defaultTemperature = temperatures["warm"]

func (s *Service) buildLights(ctx context.Context) (map[string]*Light, error) {
	lights, err := list[Light](ctx, s.req, "light")
	if err != nil {
		return nil, fmt.Errorf("list lights: %w", err)
	}

	output := make(map[string]*Light, len(lights))
	for _, light := range lights {
		light.IDV1 = strings.TrimPrefix(light.IDV1, "/lights/")

		output[light.ID] = &light
	}

	return output, err
}

func (s *Service) setWhiteLight(ctx context.Context, id, room string) error {
	var color Color
	color.XY.X = 0.372
	color.XY.Y = 0.377

	colorTemperature := ColorTemperature{
		Mirek: temperatures[s.config.Temperatures[room]],
	}

	if colorTemperature.Mirek == 0 {
		slog.LogAttrs(ctx, slog.LevelWarn, "Using default color temperature", slog.String("id", id), slog.String("room", room))
		colorTemperature.Mirek = defaultTemperature
	}

	payload := map[string]any{
		"color":             color,
		"color_temperature": colorTemperature,
	}

	if _, err := s.req.Method(http.MethodPut).Path("/clip/v2/resource/light/"+id).JSON(ctx, payload); err != nil {
		return fmt.Errorf("update light `%s`: %w", id, err)
	}

	return nil
}
