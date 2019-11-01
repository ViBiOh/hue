package hue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/iot/pkg/hue"
)

func (a *App) getLight(ctx context.Context, lightID string) (*hue.Light, error) {
	body, _, _, err := request.Get(ctx, fmt.Sprintf("%s/lights/%s", a.bridgeURL, lightID), nil)
	if err != nil {
		return nil, err
	}

	content, err := request.ReadContent(body)
	if err != nil {
		return nil, err
	}

	var light hue.Light
	if err := json.Unmarshal(content, &light); err != nil {
		return nil, err
	}

	return &light, nil
}
