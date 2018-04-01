package hue

import (
	"fmt"

	"github.com/ViBiOh/iot/pkg/hue"
)

func (a *App) listScenes() (map[string]*hue.Scene, error) {
	var response map[string]*hue.Scene

	if err := get(fmt.Sprintf(`%s/scenes`, a.bridgeURL), &response); err != nil {
		return nil, err
	}

	for id := range response {
		scene, err := a.getScene(id)
		if err != nil {
			return nil, fmt.Errorf(`Error while fetching scene %s: %v`, id, err)
		}

		response[id] = scene
	}

	return response, nil
}

func (a *App) getScene(id string) (*hue.Scene, error) {
	var response hue.Scene
	if err := get(fmt.Sprintf(`%s/scenes/%s`, a.bridgeURL, id), &response); err != nil {
		return nil, err
	}

	response.ID = id

	return &response, nil
}

func (a *App) createScene(o *hue.Scene) error {
	id, err := create(fmt.Sprintf(`%s/scenes`, a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *App) createSceneFromScheduleConfig(config *hue.ScheduleConfig, groups map[string]*hue.Group) (*hue.Scene, error) {
	group, ok := groups[config.Group]
	if !ok {
		return nil, fmt.Errorf(`Unknown group id: %s`, config.Group)
	}

	state, ok := hue.States[config.State]
	if !ok {
		return nil, fmt.Errorf(`Unknown state name: %s`, config.State)
	}

	scene := &hue.Scene{
		APIScene: &hue.APIScene{
			Name:    config.Name,
			Lights:  group.Lights,
			Recycle: false,
		},
	}

	if err := a.createScene(scene); err != nil {
		return nil, fmt.Errorf(`Error while creating scene for config %+v: %v`, config, err)
	}

	for _, light := range scene.Lights {
		if err := a.updateSceneLightState(scene, light, state); err != nil {
			return nil, fmt.Errorf(`Error while updating scene light state for config %+v: %v`, config, err)
		}
	}

	return scene, nil
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
