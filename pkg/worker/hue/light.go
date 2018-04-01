package hue

import (
	"encoding/json"
	"fmt"

	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/iot/pkg/hue"
)

func (a *App) getLight(lightID string) (*hue.Light, error) {
	content, err := request.Get(fmt.Sprintf(`%s/lights/%s`, a.bridgeURL, lightID), nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting light: %v`, err)
	}

	var light hue.Light
	if err := json.Unmarshal(content, &light); err != nil {
		return nil, fmt.Errorf(`Error while parsing light data %s: %v`, content, err)
	}

	return &light, nil
}
