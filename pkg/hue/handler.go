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
func (a *App) Handler() http.Handler {
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

func (a *App) handleGroup(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		a.rendererApp.Error(w, r, nil, model.WrapNotFound(fmt.Errorf("invalid method for updating group")))
		return
	}

	groupID := strings.Trim(strings.TrimPrefix(r.URL.Path, groupsPath), "/")
	stateName := r.FormValue("state")

	state, ok := States[stateName]
	if !ok {
		a.rendererApp.Error(w, r, nil, model.WrapNotFound(fmt.Errorf("unknown state '%s'", stateName)))
		return
	}

	group, err := a.v2App.UpdateGroup(r.Context(), groupID, state.On, float64(state.Brightness), state.Duration)
	if err != nil {
		a.rendererApp.Error(w, r, nil, err)
		return
	}

	a.rendererApp.Redirect(w, r, "/", renderer.NewSuccessMessage(fmt.Sprintf(updateSuccessMessage, group.Name, stateName)))
}

func (a *App) handleSchedule(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		a.rendererApp.Error(w, r, nil, model.WrapMethodNotAllowed(fmt.Errorf("invalid method for updating schedule")))
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
		a.rendererApp.Error(w, r, nil, err)
		return
	}

	if err := a.syncSchedules(); err != nil {
		a.rendererApp.Error(w, r, nil, err)
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

func (a *App) handleSensors(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		a.rendererApp.Error(w, r, nil, model.WrapMethodNotAllowed(fmt.Errorf("invalid method for updating sensor")))
		return
	}

	id := strings.Trim(strings.TrimPrefix(r.URL.Path, sensorsPath), "/")

	status := r.FormValue("on")
	statusBool, err := strconv.ParseBool(status)
	if err != nil {
		a.rendererApp.Error(w, r, nil, model.WrapInvalid(fmt.Errorf("unable to parse boolean with value `%s`: %s", status, err)))
		return
	}

	motionSensor, err := a.v2App.UpdateSensor(r.Context(), id, statusBool)
	if err != nil {
		a.rendererApp.Error(w, r, nil, fmt.Errorf("unable to update sensor `%s`: %s", id, err))
		return
	}

	name := motionSensor.Name + " Sensor"

	stateName := "on"
	if !statusBool {
		stateName = "off"
	}

	a.rendererApp.Redirect(w, r, "/", renderer.NewSuccessMessage(fmt.Sprintf(updateSuccessMessage, name, stateName)))
}
