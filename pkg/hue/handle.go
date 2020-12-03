package hue

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/hue/pkg/model"
)

const (
	groupsPath    = "/groups"
	schedulesPath = "/schedules"
	sensorsPath   = "/sensors"
)

// Handler for request. Should be use with net/http
func (a *app) Handle(r *http.Request) (model.Message, int) {
	if strings.HasPrefix(r.URL.Path, groupsPath) {
		return a.handleGroup(r)
	}

	if strings.HasPrefix(r.URL.Path, schedulesPath) {
		return a.handleSchedule(r)
	}

	if strings.HasPrefix(r.URL.Path, sensorsPath) {
		return a.handleSensors(r)
	}

	return emptyMessage, http.StatusOK
}

func (a *app) handleGroup(r *http.Request) (model.Message, int) {
	if r.FormValue("method") != http.MethodPatch {
		return model.NewErrorMessage("Invalid method for updating group"), http.StatusMethodNotAllowed
	}

	groupID := strings.Trim(strings.TrimPrefix(r.URL.Path, groupsPath), "/")
	stateName := r.FormValue("state")

	group, ok := a.groups[groupID]
	if !ok {
		return model.NewErrorMessage(fmt.Sprintf("unknown group '%s'", groupID)), http.StatusNotFound
	}

	state, ok := States[stateName]
	if !ok {
		return model.NewErrorMessage(fmt.Sprintf("unknown state '%s'", stateName)), http.StatusNotFound
	}

	if err := a.updateGroupState(r.Context(), groupID, state); err != nil {
		return model.NewErrorMessage(err.Error()), http.StatusInternalServerError
	}

	if err := a.syncGroups(); err != nil {
		return model.NewErrorMessage(err.Error()), http.StatusInternalServerError
	}

	return model.NewSuccessMessage(fmt.Sprintf("%s is now %s", group.Name, stateName)), http.StatusOK
}

func (a *app) handleSchedule(r *http.Request) (model.Message, int) {
	if r.FormValue("method") != http.MethodPatch {
		return model.NewErrorMessage("Invalid method for updating schedule"), http.StatusMethodNotAllowed
	}

	status := r.FormValue("status")

	schedule := Schedule{
		ID: strings.Trim(strings.TrimPrefix(r.URL.Path, schedulesPath), "/"),
		APISchedule: APISchedule{
			Status: status,
		},
	}

	if err := a.updateSchedule(r.Context(), schedule); err != nil {
		return model.NewErrorMessage(err.Error()), http.StatusInternalServerError
	}

	if err := a.syncSchedules(); err != nil {
		return model.NewErrorMessage(err.Error()), http.StatusInternalServerError
	}

	a.mutex.RLock()

	name := "Schedule"
	if updated, ok := a.schedules[schedule.ID]; ok {
		name = updated.Name
	}

	a.mutex.RUnlock()

	return model.NewSuccessMessage(fmt.Sprintf("%s is now %s", name, status)), http.StatusOK
}

func (a *app) handleSensors(r *http.Request) (model.Message, int) {
	if r.FormValue("method") != http.MethodPatch {
		return model.NewErrorMessage("Invalid method for updating sensor"), http.StatusMethodNotAllowed
	}

	status := r.FormValue("on")
	statusBool, err := strconv.ParseBool(status)
	if err != nil {
		return model.NewErrorMessage(fmt.Sprintf("unable to parse boolean with value `%s`: %s", status, err)), http.StatusInternalServerError
	}

	sensor := Sensor{
		ID: strings.Trim(strings.TrimPrefix(r.URL.Path, sensorsPath), "/"),
		Config: SensorConfig{
			On: statusBool,
		},
	}

	if err := a.updateSensorConfig(r.Context(), sensor); err != nil {
		return model.NewErrorMessage(err.Error()), http.StatusInternalServerError
	}

	if err := a.syncSensors(); err != nil {
		return model.NewErrorMessage(err.Error()), http.StatusInternalServerError
	}

	a.mutex.RLock()

	name := "Sensor"
	for _, s := range a.sensors {
		if s.ID == sensor.ID {
			name = s.Name
			break
		}
	}

	a.mutex.RUnlock()

	stateName := "on"
	if !statusBool {
		stateName = "off"
	}

	return model.NewSuccessMessage(fmt.Sprintf("%s is now %s", name, stateName)), http.StatusOK
}
