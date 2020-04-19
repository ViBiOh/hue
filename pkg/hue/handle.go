package hue

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/hue/pkg/model"
)

const (
	groupsRequest    = "/groups"
	schedulesRequest = "/schedules"
)

// Handler for request. Should be use with net/http
func (a *app) Handle(r *http.Request) (model.Message, int) {
	if strings.HasPrefix(r.URL.Path, groupsRequest) {
		return a.handleGroup(r)
	}

	if strings.HasPrefix(r.URL.Path, schedulesRequest) {
		return a.handleSchedule(r)
	}

	return emptyMessage, http.StatusOK
}

func (a *app) handleSchedule(r *http.Request) (model.Message, int) {
	if r.FormValue("method") != http.MethodPatch {
		return model.NewErrorMessage("Invalid method for updating schedule"), http.StatusMethodNotAllowed
	}

	status := r.FormValue("status")

	schedule := Schedule{
		ID: strings.Trim(strings.TrimPrefix(r.URL.Path, schedulesRequest), "/"),
		APISchedule: APISchedule{
			Status: r.FormValue("status"),
		},
	}

	if err := a.updateSchedule(r.Context(), schedule); err != nil {
		return model.NewErrorMessage(err.Error()), http.StatusInternalServerError
	}

	if err := a.syncSchedules(); err != nil {
		return model.NewErrorMessage(err.Error()), http.StatusInternalServerError
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	name := "Schedule"
	if updated, ok := a.schedules[schedule.ID]; ok {
		name = updated.Name
	}

	return model.NewSuccessMessage(fmt.Sprintf("%s is now %s", name, status)), http.StatusOK
}

func (a *app) handleGroup(r *http.Request) (model.Message, int) {
	if r.FormValue("method") != http.MethodPatch {
		return model.NewErrorMessage("Invalid method for updating group"), http.StatusMethodNotAllowed
	}

	groupID := strings.Trim(strings.TrimPrefix(r.URL.Path, groupsRequest), "/")
	stateName := r.FormValue("state")

	group, ok := a.groups[groupID]
	if !ok {
		return model.NewErrorMessage(fmt.Sprintf("Unknown group '%s'", groupID)), http.StatusNotFound
	}

	state, ok := States[stateName]
	if !ok {
		return model.NewErrorMessage(fmt.Sprintf("Unknown state '%s'", stateName)), http.StatusNotFound
	}

	if err := a.updateGroupState(r.Context(), groupID, state); err != nil {
		return model.NewErrorMessage(err.Error()), http.StatusInternalServerError
	}

	if err := a.syncGroups(); err != nil {
		return model.NewErrorMessage(err.Error()), http.StatusInternalServerError
	}

	return model.NewSuccessMessage(fmt.Sprintf("%s is now %s", group.Name, stateName)), http.StatusOK
}
