package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils"
)

func (a *App) listScenes() (map[string]interface{}, error) {
	content, err := httputils.GetRequest(a.bridgeURL+`/scenes`, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting scenes: %v`, err)
	}

	var scenes map[string]interface{}
	if err := json.Unmarshal(content, &scenes); err != nil {
		return nil, fmt.Errorf(`Error while parsing scenes: %v`, err)
	}

	return scenes, nil
}

func (a *App) deleteScene(id string) error {
	content, err := httputils.Request(a.bridgeURL+`/scenes/`+id, nil, nil, http.MethodDelete)
	if err != nil {
		return fmt.Errorf(`Error while deleting scene: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while deleting scene: %s`, content)
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
