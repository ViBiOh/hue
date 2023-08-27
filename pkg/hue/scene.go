package hue

import (
	"context"
	"fmt"
)

func (s *Service) listScenes(ctx context.Context) (map[string]Scene, error) {
	var response map[string]Scene

	if err := get(ctx, fmt.Sprintf("%s/scenes", s.bridgeURL), &response); err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	for id := range response {
		scene, err := s.getScene(ctx, id)
		if err != nil {
			return nil, err
		}

		response[id] = scene
	}

	return response, nil
}

func (s *Service) getScene(ctx context.Context, id string) (Scene, error) {
	var response Scene
	if err := get(ctx, fmt.Sprintf("%s/scenes/%s", s.bridgeURL, id), &response); err != nil {
		return response, err
	}

	response.ID = id

	return response, nil
}

func (s *Service) createScene(ctx context.Context, o *Scene) error {
	id, err := create(ctx, fmt.Sprintf("%s/scenes", s.bridgeURL), o)
	if err != nil {
		return err
	}

	o.ID = id

	return nil
}

func (s *Service) createSceneFromScheduleConfig(ctx context.Context, config ScheduleConfig, groups map[string]Group) (Scene, error) {
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

	if err := s.createScene(ctx, &scene); err != nil {
		return scene, err
	}

	for _, light := range scene.Lights {
		if err := s.updateSceneLightState(ctx, scene, light, state); err != nil {
			return scene, err
		}
	}

	return scene, nil
}

func (s *Service) updateSceneLightState(ctx context.Context, o Scene, lightID string, state State) error {
	return update(ctx, fmt.Sprintf("%s/scenes/%s/lightstates/%s", s.bridgeURL, o.ID, lightID), state.V1())
}

func (s *Service) deleteScene(ctx context.Context, id string) error {
	return remove(ctx, fmt.Sprintf("%s/scenes/%s", s.bridgeURL, id))
}

func (s *Service) cleanScenes(ctx context.Context) error {
	scenes, err := s.listScenes(ctx)
	if err != nil {
		return err
	}

	for key := range scenes {
		if err := s.deleteScene(ctx, key); err != nil {
			return err
		}
	}

	return nil
}
