package hue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ViBiOh/httputils/v3/pkg/request"
)

func (a *app) getLight(ctx context.Context, lightID string) (*Light, error) {
	resp, err := request.New().Get(fmt.Sprintf("%s/lights/%s", a.bridgeURL, lightID)).Send(ctx, nil)
	if err != nil {
		return nil, err
	}

	content, err := request.ReadBodyResponse(resp)
	if err != nil {
		return nil, err
	}

	var light Light
	if err := json.Unmarshal(content, &light); err != nil {
		return nil, err
	}

	return &light, nil
}
