package hue

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
)

const (
	apiPath       = "/api"
	groupsPath    = "/groups"
	schedulesPath = "/schedules"
	sensorsPath   = "/sensors"

	updateSuccessMessage = "%s is now %s"
)

// Handler for request. Should be use with net/http
func (a *app) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, groupsPath) {
			a.handleGroup(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, schedulesPath) {
			a.handleSchedule(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, sensorsPath) {
			a.handleSensors(w, r)
			return
		}

		httperror.NotFound(w)
	})
}

func (a *app) handleGroup(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		a.rendererApp.Error(w, model.WrapNotFound(fmt.Errorf("invalid method for updating group")))
		return
	}

	groupID := strings.Trim(strings.TrimPrefix(r.URL.Path, groupsPath), "/")
	stateName := r.FormValue("state")

	group, ok := a.groups[groupID]
	if !ok {
		a.rendererApp.Error(w, model.WrapNotFound(fmt.Errorf("unknown group '%s'", groupID)))
		return
	}

	state, ok := States[stateName]
	if !ok {
		a.rendererApp.Error(w, model.WrapNotFound(fmt.Errorf("unknown state '%s'", stateName)))
		return
	}

	if err := a.updateGroupState(r.Context(), groupID, state); err != nil {
		a.rendererApp.Error(w, err)
		return
	}

	if err := a.syncGroups(); err != nil {
		a.rendererApp.Error(w, err)
		return
	}

	a.rendererApp.Redirect(w, r, "/", renderer.NewSuccessMessage(fmt.Sprintf(updateSuccessMessage, group.Name, stateName)))
}

func (a *app) handleSchedule(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		a.rendererApp.Error(w, model.WrapMethodNotAllowed(fmt.Errorf("invalid method for updating schedule")))
		return
	}

	status := r.FormValue("status")

	schedule := Schedule{
		ID: strings.Trim(strings.TrimPrefix(r.URL.Path, schedulesPath), "/"),
		APISchedule: APISchedule{
			Status: status,
		},
	}

	if err := a.updateSchedule(r.Context(), schedule); err != nil {
		a.rendererApp.Error(w, err)
		return
	}

	if err := a.syncSchedules(); err != nil {
		a.rendererApp.Error(w, err)
		return
	}

	a.mutex.RLock()

	name := "Schedule"
	if updated, ok := a.schedules[schedule.ID]; ok {
		name = updated.Name
	}

	a.mutex.RUnlock()

	a.rendererApp.Redirect(w, r, "/", renderer.NewSuccessMessage(fmt.Sprintf(updateSuccessMessage, name, status)))
}

func (a *app) handleSensors(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		a.rendererApp.Error(w, model.WrapMethodNotAllowed(fmt.Errorf("invalid method for updating sensor")))
		return
	}

	status := r.FormValue("on")
	statusBool, err := strconv.ParseBool(status)
	if err != nil {
		a.rendererApp.Error(w, model.WrapInvalid(fmt.Errorf("unable to parse boolean with value `%s`: %s", status, err)))
		return
	}

	sensor := Sensor{
		ID: strings.Trim(strings.TrimPrefix(r.URL.Path, sensorsPath), "/"),
		Config: SensorConfig{
			On: statusBool,
		},
	}

	if err := a.updateSensorConfig(r.Context(), sensor); err != nil {
		a.rendererApp.Error(w, err)
		return
	}

	if err := a.syncSensors(); err != nil {
		a.rendererApp.Error(w, err)
		return
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

	a.rendererApp.Redirect(w, r, "/", renderer.NewSuccessMessage(fmt.Sprintf(updateSuccessMessage, name, stateName)))
}
