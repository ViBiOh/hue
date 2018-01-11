package hue

import (
	"fmt"

	"github.com/ViBiOh/iot/hue"
)

func (a *App) listScenes() (map[string]*hue.Scene, error) {
	var response map[string]*hue.Scene
	return response, get(fmt.Sprintf(`%s/scenes`, a.bridgeURL), &response)
}

func (a *App) createScene(o *hue.Scene) error {
	id, err := create(fmt.Sprintf(`%s/scenes`, a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *App) updateSceneLightState(o *hue.Scene, lightID string, state map[string]interface{}) error {
	return update(fmt.Sprintf(`%s/scenes/%s/lightstates/%s`, a.bridgeURL, o.ID, lightID), state)
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
