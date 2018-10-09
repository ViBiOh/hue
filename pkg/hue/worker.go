package hue

import (
	"encoding/json"
	"fmt"

	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	// WorkerGroupsAction ws message prefix for groups command
	WorkerGroupsAction = `groups`

	// WorkerSchedulesAction ws message prefix for schedules command
	WorkerSchedulesAction = `schedules`

	// WorkerScenesAction ws message prefix for scenes command
	WorkerScenesAction = `scenes`

	// WorkerStateAction ws message prefix for state command
	WorkerStateAction = `state`

	// CreateAction ws message prefix for create command
	CreateAction = `create`

	// UpdateAction ws message prefix for update command
	UpdateAction = `update`

	// DeleteAction ws message prefix for delete command
	DeleteAction = `delete`
)

func (a *App) handleGroupsFromWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var newGroups map[string]*Group
	if err := json.Unmarshal([]byte(message.Payload), &newGroups); err != nil {
		return fmt.Errorf(`error while unmarshalling groups: %v`, err)
	}

	a.groups = newGroups

	return nil
}

func (a *App) handleSchedulesFromWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var newSchedules map[string]*Schedule
	if err := json.Unmarshal([]byte(message.Payload), &newSchedules); err != nil {
		return fmt.Errorf(`error while unmarshalling schedules: %v`, err)
	}

	a.schedules = newSchedules

	return nil
}

func (a *App) handleScenesFromWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var newScenes map[string]*Scene
	if err := json.Unmarshal([]byte(message.Payload), &newScenes); err != nil {
		return fmt.Errorf(`error while unmarshalling scenes: %v`, err)
	}

	a.scenes = newScenes

	return nil
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(p *provider.WorkerMessage) error {
	if p.Action == WorkerGroupsAction {
		return a.handleGroupsFromWorker(p)
	}

	if p.Action == WorkerSchedulesAction {
		return a.handleSchedulesFromWorker(p)
	}

	if p.Action == WorkerScenesAction {
		return a.handleScenesFromWorker(p)
	}

	return provider.ErrWorkerUnknownAction
}
