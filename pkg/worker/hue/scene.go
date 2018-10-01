package hue

import (
	"context"
	"fmt"

	"github.com/ViBiOh/iot/pkg/hue"
)

func (a *App) listScenes(ctx context.Context) (map[string]*hue.Scene, error) {
	var response map[string]*hue.Scene

	if err := get(ctx, fmt.Sprintf(`%s/scenes`, a.bridgeURL), &response); err != nil {
		return nil, err
	}

	for id := range response {
		scene, err := a.getScene(ctx, id)
		if err != nil {
			return nil, fmt.Errorf(`error while fetching scene %s: %v`, id, err)
		}

		response[id] = scene
	}

	return response, nil
}

func (a *App) getScene(ctx context.Context, id string) (*hue.Scene, error) {
	var response hue.Scene
	if err := get(ctx, fmt.Sprintf(`%s/scenes/%s`, a.bridgeURL, id), &response); err != nil {
		return nil, err
	}

	response.ID = id

	return &response, nil
}

func (a *App) createScene(ctx context.Context, o *hue.Scene) error {
	id, err := create(ctx, fmt.Sprintf(`%s/scenes`, a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *App) createSceneFromScheduleConfig(ctx context.Context, config *hue.ScheduleConfig, groups map[string]*hue.Group) (*hue.Scene, error) {
	group, ok := groups[config.Group]
	if !ok {
		return nil, fmt.Errorf(`unknown group id: %s`, config.Group)
	}

	state, ok := hue.States[config.State]
	if !ok {
		return nil, fmt.Errorf(`unknown state name: %s`, config.State)
	}

	scene := &hue.Scene{
		APIScene: &hue.APIScene{
			Name:    config.Name,
			Lights:  group.Lights,
			Recycle: false,
		},
	}

	if err := a.createScene(ctx, scene); err != nil {
		return nil, fmt.Errorf(`error while creating scene for config %+v: %v`, config, err)
	}

	for _, light := range scene.Lights {
		if err := a.updateSceneLightState(ctx, scene, light, state); err != nil {
			return nil, fmt.Errorf(`error while updating scene light state for config %+v: %v`, config, err)
		}
	}

	return scene, nil
}

func (a *App) updateSceneLightState(ctx context.Context, o *hue.Scene, lightID string, state map[string]interface{}) error {
	return update(ctx, fmt.Sprintf(`%s/scenes/%s/lightstates/%s`, a.bridgeURL, o.ID, lightID), state)
}

func (a *App) deleteScene(ctx context.Context, id string) error {
	return delete(ctx, fmt.Sprintf(`%s/scenes/%s`, a.bridgeURL, id))
}

func (a *App) cleanScenes(ctx context.Context) error {
	scenes, err := a.listScenes(ctx)
	if err != nil {
		return fmt.Errorf(`error while listing scenes: %v`, err)
	}

	for key := range scenes {
		if err := a.deleteScene(ctx, key); err != nil {
			return fmt.Errorf(`error while deleting scene: %v`, err)
		}
	}

	return nil
}
