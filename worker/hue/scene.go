package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils"
)

func (a *App) listScenes() (map[string]interface{}, error) {
	content, err := httputils.GetRequest(fmt.Sprintf(`%s/scenes`, a.bridgeURL), nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while sending get request: %v`, err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(content, &response); err != nil {
		return nil, fmt.Errorf(`Error while parsing response: %v`, err)
	}

	return response, nil
}

func (a *App) deleteScene(id string) error {
	content, err := httputils.Request(fmt.Sprintf(`%s/scenes/%s`, a.bridgeURL, id), nil, nil, http.MethodDelete)
	if err != nil {
		return fmt.Errorf(`Error while sending delete request: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while sending delete request: %s`, content)
	}

	return nil
}

func (a *App) cleanScenes() error {
	scenes, err := a.listScenes()
	if err != nil {
		return fmt.Errorf(`Error while listing scenes: %v`, err)
	}

	for key := range scenes {
		if err := a.deleteScene(key); err != nil {
			return fmt.Errorf(`Error while deleting scene: %v`, err)
		}
	}

	return nil
}
