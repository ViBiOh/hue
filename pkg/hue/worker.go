package hue

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	// WorkerGroupsType ws message prefix for groups command
	WorkerGroupsType = `groups`

	// WorkerSchedulesType ws message prefix for schedules command
	WorkerSchedulesType = `schedules`

	// WorkerScenesType ws message prefix for scenes command
	WorkerScenesType = `scenes`

	// WorkerStateType ws message prefix for state command
	WorkerStateType = `state`

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
		return fmt.Errorf(`Error while converting groups payload: %v`, err)
	}

	if err := json.Unmarshal(convert, &newGroups); err != nil {
		return fmt.Errorf(`Error while unmarshalling groups: %v`, err)
	}

	a.groups = newGroups

	return nil
}

func (a *App) handleSchedulesFromWorker(message *provider.WorkerMessage) error {
	var newSchedules map[string]*Schedule

	convert, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf(`Error while converting groups payload: %v`, err)
	}

	if err := json.Unmarshal(convert, &newSchedules); err != nil {
		return fmt.Errorf(`Error while unmarshalling schedules: %v`, err)
	}

	a.schedules = newSchedules

	return nil
}

func (a *App) handleScenesFromWorker(message *provider.WorkerMessage) error {
	var newScenes map[string]*Scene

	convert, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf(`Error while converting groups payload: %v`, err)
	}

	if err := json.Unmarshal(convert, &newScenes); err != nil {
		return fmt.Errorf(`Error while unmarshalling scenes: %v`, err)
	}

	a.scenes = newScenes

	return nil
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(message *provider.WorkerMessage) error {
	if strings.HasPrefix(message.Type, WorkerGroupsType) {
		return a.handleGroupsFromWorker(message)
	}

	if strings.HasPrefix(message.Type, WorkerSchedulesType) {
		return a.handleSchedulesFromWorker(message)
	}

	if strings.HasPrefix(message.Type, WorkerScenesType) {
		return a.handleScenesFromWorker(message)
	}

	return errors.New(`Unknown command`)
}
