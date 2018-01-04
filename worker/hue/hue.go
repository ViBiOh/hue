package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/iot/hue"
)

// GetURL forge bridge URL
func GetURL(bridgeIP, username string) string {
	return `http://` + bridgeIP + `/api/` + username + `/lights`
}

func listLights(bridgeURL string) ([]hue.Light, error) {
	content, err := httputils.GetBody(bridgeURL, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting data from bridge: %v`, err)
	}

	var rawLights map[string]hue.Light
	if err := json.Unmarshal(content, &rawLights); err != nil {
		return nil, fmt.Errorf(`Error while parsing data from bridge: %v`, err)
	}

	lights := make([]hue.Light, len(rawLights))
	for key, value := range rawLights {
		i, _ := strconv.Atoi(key)
		lights[i-1] = value
	}

	return lights, nil
}

// ListLightsJSON get lists of lights in JSON
func ListLightsJSON(bridgeURL string) ([]byte, error) {
	lights, err := listLights(bridgeURL)
	if err != nil {
		err = fmt.Errorf(`Error while listing lights: %v`, err)
		return nil, err
	}

	lightsJSON, err := json.Marshal(lights)
	if err != nil {
		err = fmt.Errorf(`Error while marshalling lights: %v`, err)
		return nil, err
	}

	return lightsJSON, nil
}

func updateState(bridgeURL, light, state string) error {
	content, err := httputils.MethodBody(bridgeURL+`/`+light+`/state`, []byte(state), nil, http.MethodPut)

	if err != nil {
		return fmt.Errorf(`Error while sending data to bridge: %v`, err)
	}

	if bytes.Contains(content, []byte(`error`)) {
		return fmt.Errorf(`Error while updating state: %s`, content)
	}

	return nil
}

// UpdateAllState updates state of all lights
func UpdateAllState(bridgeURL, state string) error {
	lights, err := listLights(bridgeURL)
	if err != nil {
		return fmt.Errorf(`Error while listing lights: %v`, err)
	}

	for index, light := range lights {
		if err := updateState(bridgeURL, strconv.Itoa(index+1), state); err != nil {
			return fmt.Errorf(`Error while updating %s: %v`, light.Name, err)
		}
	}

	return nil
}
