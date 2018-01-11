package hue

import (
	"fmt"
)

func (a *App) listScenes() (map[string]interface{}, error) {
	var response map[string]interface{}
	return response, get(fmt.Sprintf(`%s/scenes`, a.bridgeURL), response)
}

func (a *App) deleteScene(id string) error {
	return delete(fmt.Sprintf(`%s/scenes/%s`, a.bridgeURL, id))
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
