package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
)

var (
	// WebSocketPrefix ws message prefix for all hue commands
	WebSocketPrefix = []byte(`hue `)

	// GroupsPrefix ws message prefix for groups command
	GroupsPrefix = []byte(`groups `)

	// SchedulesPrefix ws message prefix for schedules command
	SchedulesPrefix = []byte(`schedules `)

	// ScenesPrefix ws message prefix for scenes command
	ScenesPrefix = []byte(`scenes `)

	// StatePrefix ws message prefix for state command
	StatePrefix = []byte(`state `)

	// CreatePrefix ws message prefix for create command
	CreatePrefix = []byte(`create `)

	// UpdatePrefix ws message prefix for update command
	UpdatePrefix = []byte(`update `)

	// DeletePrefix ws message prefix for delete command
	DeletePrefix = []byte(`delete `)
)

func (a *App) handleGroupsFromWorker(payload []byte) error {
	var newGroups map[string]*Group

	if err := json.Unmarshal(payload, &newGroups); err != nil {
		return fmt.Errorf(`[hue] Error while unmarshalling groups: %v`, err)
	}

	a.groups = newGroups

	return nil
}

func (a *App) handleSchedulesFromWorker(payload []byte) error {
	var newSchedules map[string]*Schedule

	if err := json.Unmarshal(payload, &newSchedules); err != nil {
		return fmt.Errorf(`[hue] Error while unmarshalling schedules: %v`, err)
	}

	a.schedules = newSchedules

	return nil
}

func (a *App) handleScenesFromWorker(payload []byte) error {
	var newScenes map[string]*Scene

	if err := json.Unmarshal(bytes.TrimPrefix(payload, SchedulesPrefix), &newScenes); err != nil {
		return fmt.Errorf(`[hue] Error while unmarshalling scenes: %v`, err)
	}

	a.scenes = newScenes

	return nil
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(payload []byte) error {
	if bytes.HasPrefix(payload, GroupsPrefix) {
		return a.handleGroupsFromWorker(bytes.TrimPrefix(payload, GroupsPrefix))
	}

	if bytes.HasPrefix(payload, SchedulesPrefix) {
		return a.handleSchedulesFromWorker(bytes.TrimPrefix(payload, SchedulesPrefix))
	}

	if bytes.HasPrefix(payload, ScenesPrefix) {
		return a.handleScenesFromWorker(bytes.TrimPrefix(payload, ScenesPrefix))
	}

	return fmt.Errorf(`[hue] Unknown command`)
}
