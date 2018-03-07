package hue

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ViBiOh/iot/provider"
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

	// CreatePrefix ws message prefix for create command
	CreatePrefix = `create`

	// UpdatePrefix ws message prefix for update command
	UpdatePrefix = `update`

	// DeletePrefix ws message prefix for delete command
	DeletePrefix = `delete`
)

func (a *App) handleGroupsFromWorker(message *provider.WorkerMessage) error {
	var newGroups map[string]*Group

	if convert, err := json.Marshal(message.Payload); err != nil {
		return fmt.Errorf(`[%s] Error while converting groups payload: %v`, HueSource, err)
	} else if err := json.Unmarshal(convert, &newGroups); err != nil {
		return fmt.Errorf(`[%s] Error while unmarshalling groups: %v`, HueSource, err)
	}

	a.groups = newGroups

	return nil
}

func (a *App) handleSchedulesFromWorker(message *provider.WorkerMessage) error {
	var newSchedules map[string]*Schedule

	if convert, err := json.Marshal(message.Payload); err != nil {
		return fmt.Errorf(`[%s] Error while converting groups payload: %v`, HueSource, err)
	} else if err := json.Unmarshal(convert, &newSchedules); err != nil {
		return fmt.Errorf(`[%s] Error while unmarshalling schedules: %v`, HueSource, err)
	}

	a.schedules = newSchedules

	return nil
}

func (a *App) handleScenesFromWorker(message *provider.WorkerMessage) error {
	var newScenes map[string]*Scene

	if convert, err := json.Marshal(message.Payload); err != nil {
		return fmt.Errorf(`[%s] Error while converting groups payload: %v`, HueSource, err)
	} else if err := json.Unmarshal(convert, &newScenes); err != nil {
		return fmt.Errorf(`[%s] Error while unmarshalling scenes: %v`, HueSource, err)
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

	return fmt.Errorf(`[%s] Unknown command`, HueSource)
}
