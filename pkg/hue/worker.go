package hue

import (
	"encoding/json"

	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	// WorkerGroupsAction message prefix for groups command
	WorkerGroupsAction = "groups"

	// WorkerSchedulesAction message prefix for schedules command
	WorkerSchedulesAction = "schedules"

	// WorkerScenesAction message prefix for scenes command
	WorkerScenesAction = "scenes"

	// WorkerSensorsAction message prefix for sensors command
	WorkerSensorsAction = "sensors"

	// WorkerStateAction message prefix for state command
	WorkerStateAction = "state"

	// CreateAction message prefix for create command
	CreateAction = "create"

	// UpdateAction message prefix for update command
	UpdateAction = "update"

	// DeleteAction message prefix for delete command
	DeleteAction = "delete"
)

func (a *app) handleGroupsFromWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var newGroups map[string]*Group
	if err := json.Unmarshal([]byte(message.Payload), &newGroups); err != nil {
		return err
	}

	a.groups = newGroups

	return nil
}

func (a *app) handleSchedulesFromWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var newSchedules map[string]*Schedule
	if err := json.Unmarshal([]byte(message.Payload), &newSchedules); err != nil {
		return err
	}

	a.schedules = newSchedules

	return nil
}

func (a *app) handleSensorsFromWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var newSensors map[string]*Sensor
	if err := json.Unmarshal([]byte(message.Payload), &newSensors); err != nil {
		return err
	}

	a.sensors = newSensors

	a.updatePrometheusSensors()

	return nil
}

func (a *app) handleScenesFromWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var newScenes map[string]*Scene
	if err := json.Unmarshal([]byte(message.Payload), &newScenes); err != nil {
		return err
	}

	a.scenes = newScenes

	return nil
}

// WorkerHandler handle commands receive from worker
func (a *app) WorkerHandler(p *provider.WorkerMessage) error {
	if p.Action == WorkerGroupsAction {
		return a.handleGroupsFromWorker(p)
	}

	if p.Action == WorkerSchedulesAction {
		return a.handleSchedulesFromWorker(p)
	}

	if p.Action == WorkerSensorsAction {
		return a.handleSensorsFromWorker(p)
	}

	if p.Action == WorkerScenesAction {
		return a.handleScenesFromWorker(p)
	}

	return provider.ErrWorkerUnknownAction
}
