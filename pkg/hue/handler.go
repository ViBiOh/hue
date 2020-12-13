package hue

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/renderer/model"
)

const (
	groupsPath    = "/groups"
	schedulesPath = "/schedules"
	sensorsPath   = "/sensors"
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
		a.renderer.Error(w, model.WrapNotFound(fmt.Errorf("Invalid method for updating group")))
		return
	}

	groupID := strings.Trim(strings.TrimPrefix(r.URL.Path, groupsPath), "/")
	stateName := r.FormValue("state")

	group, ok := a.groups[groupID]
	if !ok {
		a.renderer.Error(w, model.WrapNotFound(fmt.Errorf("unknown group '%s'", groupID)))
		return
	}

	state, ok := States[stateName]
	if !ok {
		a.renderer.Error(w, model.WrapNotFound(fmt.Errorf("unknown state '%s'", stateName)))
		return
	}

	if err := a.updateGroupState(r.Context(), groupID, state); err != nil {
		a.renderer.Error(w, err)
		return
	}

	if err := a.syncGroups(); err != nil {
		a.renderer.Error(w, err)
		return
	}

	a.renderer.Redirect(w, r, "/", fmt.Sprintf("%s is now %s", group.Name, stateName))
}

func (a *app) handleSchedule(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		a.renderer.Error(w, model.WrapMethodNotAllowed(fmt.Errorf("Invalid method for updating schedule")))
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
		a.renderer.Error(w, err)
		return
	}

	if err := a.syncSchedules(); err != nil {
		a.renderer.Error(w, err)
		return
	}

	a.mutex.RLock()

	name := "Schedule"
	if updated, ok := a.schedules[schedule.ID]; ok {
		name = updated.Name
	}

	a.mutex.RUnlock()

	a.renderer.Redirect(w, r, "/", fmt.Sprintf("%s is now %s", name, status))
}

func (a *app) handleSensors(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		a.renderer.Error(w, model.WrapMethodNotAllowed(fmt.Errorf("Invalid method for updating sensor")))
		return
	}

	status := r.FormValue("on")
	statusBool, err := strconv.ParseBool(status)
	if err != nil {
		a.renderer.Error(w, model.WrapInvalid(fmt.Errorf("unable to parse boolean with value `%s`: %s", status, err)))
		return
	}

	sensor := Sensor{
		ID: strings.Trim(strings.TrimPrefix(r.URL.Path, sensorsPath), "/"),
		Config: SensorConfig{
			On: statusBool,
		},
	}

	if err := a.updateSensorConfig(r.Context(), sensor); err != nil {
		a.renderer.Error(w, err)
		return
	}

	if err := a.syncSensors(); err != nil {
		a.renderer.Error(w, err)
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

	a.renderer.Redirect(w, r, "/", fmt.Sprintf("%s is now %s", name, stateName))
}
