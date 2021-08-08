package hue

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func (a *App) getLight(ctx context.Context, lightID string) (Light, error) {
	resp, err := request.New().Get(fmt.Sprintf("%s/lights/%s", a.bridgeURL, lightID)).Send(ctx, nil)
	if err != nil {
		return noneLight, err
	}

	var light Light
	if err := httpjson.Read(resp, &light); err != nil {
		return noneLight, fmt.Errorf("unable to read light payload: %s", err)
	}

	return light, nil
}
