package hue

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ViBiOh/iot/provider"
)

const (
	// GroupsPrefix ws message prefix for groups command
	GroupsPrefix = `groups`

	// SchedulesPrefix ws message prefix for schedules command
	SchedulesPrefix = `schedules`

	// ScenesPrefix ws message prefix for scenes command
	ScenesPrefix = `scenes`

	// StatePrefix ws message prefix for state command
	StatePrefix = `state`

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
		return fmt.Errorf(`[hue] Error while converting groups payload: %v`, err)
	} else if err := json.Unmarshal(convert, &newGroups); err != nil {
		return fmt.Errorf(`[hue] Error while unmarshalling groups: %v`, err)
	}

	a.groups = newGroups

	return nil
}

func (a *App) handleSchedulesFromWorker(message *provider.WorkerMessage) error {
	var newSchedules map[string]*Schedule

	if convert, err := json.Marshal(message.Payload); err != nil {
		return fmt.Errorf(`[hue] Error while converting groups payload: %v`, err)
	} else if err := json.Unmarshal(convert, &newSchedules); err != nil {
		return fmt.Errorf(`[hue] Error while unmarshalling schedules: %v`, err)
	}

	a.schedules = newSchedules

	return nil
}

func (a *App) handleScenesFromWorker(message *provider.WorkerMessage) error {
	var newScenes map[string]*Scene

	if convert, err := json.Marshal(message.Payload); err != nil {
		return fmt.Errorf(`[hue] Error while converting groups payload: %v`, err)
	} else if err := json.Unmarshal(convert, &newScenes); err != nil {
		return fmt.Errorf(`[hue] Error while unmarshalling scenes: %v`, err)
	}

	a.scenes = newScenes

	return nil
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(message *provider.WorkerMessage) error {
	if strings.HasPrefix(message.Type, GroupsPrefix) {
		return a.handleGroupsFromWorker(message)
	}

	if strings.HasPrefix(message.Type, SchedulesPrefix) {
		return a.handleSchedulesFromWorker(message)
	}

	if strings.HasPrefix(message.Type, ScenesPrefix) {
		return a.handleScenesFromWorker(message)
	}

	return fmt.Errorf(`[hue] Unknown command`)
}
