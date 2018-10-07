package hue

import (
	"encoding/json"
	"fmt"
	"strings"

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
	var newGroups map[string]*Group

	convert, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf(`error while converting groups payload: %v`, err)
	}

	if err := json.Unmarshal(convert, &newGroups); err != nil {
		return fmt.Errorf(`error while unmarshalling groups: %v`, err)
	}

	a.groups = newGroups

	return nil
}

func (a *App) handleSchedulesFromWorker(message *provider.WorkerMessage) error {
	var newSchedules map[string]*Schedule

	convert, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf(`error while converting groups payload: %v`, err)
	}

	if err := json.Unmarshal(convert, &newSchedules); err != nil {
		return fmt.Errorf(`error while unmarshalling schedules: %v`, err)
	}

	a.schedules = newSchedules

	return nil
}

func (a *App) handleScenesFromWorker(message *provider.WorkerMessage) error {
	var newScenes map[string]*Scene

	convert, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf(`error while converting groups payload: %v`, err)
	}

	if err := json.Unmarshal(convert, &newScenes); err != nil {
		return fmt.Errorf(`error while unmarshalling scenes: %v`, err)
	}

	a.scenes = newScenes

	return nil
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(message *provider.WorkerMessage) error {
	if strings.HasPrefix(message.Action, WorkerGroupsAction) {
		return a.handleGroupsFromWorker(message)
	}

	if strings.HasPrefix(message.Action, WorkerSchedulesAction) {
		return a.handleSchedulesFromWorker(message)
	}

	if strings.HasPrefix(message.Action, WorkerScenesAction) {
		return a.handleScenesFromWorker(message)
	}

	return provider.WorkerUnknownActionErr
}
