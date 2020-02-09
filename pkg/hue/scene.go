package hue

import (
	"context"
	"fmt"
)

func (a *app) listScenes(ctx context.Context) (map[string]Scene, error) {
	var response map[string]Scene

	if err := get(ctx, fmt.Sprintf("%s/scenes", a.bridgeURL), &response); err != nil {
		return nil, err
	}

	for id := range response {
		scene, err := a.getScene(ctx, id)
		if err != nil {
			return nil, err
		}

		response[id] = scene
	}

	return response, nil
}

func (a *app) getScene(ctx context.Context, id string) (Scene, error) {
	var response Scene
	if err := get(ctx, fmt.Sprintf("%s/scenes/%s", a.bridgeURL, id), &response); err != nil {
		return response, err
	}

	response.ID = id

	return response, nil
}

func (a *app) createScene(ctx context.Context, o Scene) error {
	id, err := create(ctx, fmt.Sprintf("%s/scenes", a.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = *id

	return nil
}

func (a *app) createSceneFromScheduleConfig(ctx context.Context, config ScheduleConfig, groups map[string]Group) (Scene, error) {
	group, ok := groups[config.Group]
	if !ok {
		return Scene{}, fmt.Errorf("unknown group id: %s", config.Group)
	}

	state, ok := States[config.State]
	if !ok {
		return Scene{}, fmt.Errorf("unknown state name: %s", config.State)
	}

	scene := Scene{
		APIScene: APIScene{
			Name:    config.Name,
			Lights:  group.Lights,
			Recycle: false,
		},
	}

	if err := a.createScene(ctx, scene); err != nil {
		return scene, err
	}

	for _, light := range scene.Lights {
		if err := a.updateSceneLightState(ctx, scene, light, state); err != nil {
			return scene, err
		}
	}

	return scene, nil
}

func (a *app) updateSceneLightState(ctx context.Context, o Scene, lightID string, state map[string]interface{}) error {
	return update(ctx, fmt.Sprintf("%s/scenes/%s/lightstates/%s", a.bridgeURL, o.ID, lightID), state)
}

func (a *app) deleteScene(ctx context.Context, id string) error {
	return delete(ctx, fmt.Sprintf("%s/scenes/%s", a.bridgeURL, id))
}

func (a *app) cleanScenes(ctx context.Context) error {
	scenes, err := a.listScenes(ctx)
	if err != nil {
		return err
	}

	for key := range scenes {
		if err := a.deleteScene(ctx, key); err != nil {
			return err
		}
	}

	return nil
}
