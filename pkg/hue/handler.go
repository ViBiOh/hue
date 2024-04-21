package hue

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
)

const (
	updateSuccessMessage = "%s is now %s"
)

func (s *Service) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.PathValue("resource") {
		case "groups":
			s.handleGroup(w, r)

		case "schedules":
			s.handleSchedule(w, r)

		case "sensors":
			s.handleSensors(w, r)

		default:
			httperror.NotFound(r.Context(), w)
		}
	})
}

func (s *Service) handleGroup(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		s.renderer.Error(w, r, nil, model.WrapNotFound(fmt.Errorf("invalid method for updating group")))
		return
	}

	groupID := r.PathValue("id")
	stateName := r.FormValue("state")

	state, ok := States[stateName]
	if !ok {
		s.renderer.Error(w, r, nil, model.WrapNotFound(fmt.Errorf("unknown state '%s'", stateName)))
		return
	}

	group, err := s.v2Service.UpdateGroup(r.Context(), groupID, state.On, float64(state.Brightness), state.Duration)
	if err != nil {
		s.renderer.Error(w, r, nil, err)
		return
	}

	s.renderer.Redirect(w, r, "/", renderer.NewSuccessMessage(fmt.Sprintf(updateSuccessMessage, group.Name, stateName)))
}

func (s *Service) handleSchedule(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		s.renderer.Error(w, r, nil, model.WrapMethodNotAllowed(fmt.Errorf("invalid method for updating schedule")))
		return
	}

	status := r.FormValue("status")

	schedule := Schedule{
		ID: r.PathValue("id"),
		APISchedule: APISchedule{
			Status: status,
		},
	}

	ctx := r.Context()

	if err := s.updateSchedule(ctx, schedule); err != nil {
		s.renderer.Error(w, r, nil, err)
		return
	}

	if err := s.syncSchedules(ctx); err != nil {
		s.renderer.Error(w, r, nil, err)
		return
	}

	s.mutex.RLock()

	name := "Schedule"
	if updated, ok := s.schedules[schedule.ID]; ok {
		name = updated.Name
	}

	s.mutex.RUnlock()

	s.renderer.Redirect(w, r, "/", renderer.NewSuccessMessage(fmt.Sprintf(updateSuccessMessage, name, status)))
}

func (s *Service) handleSensors(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("method") != http.MethodPatch {
		s.renderer.Error(w, r, nil, model.WrapMethodNotAllowed(fmt.Errorf("invalid method for updating sensor")))
		return
	}

	id := r.PathValue("id")

	status := r.FormValue("on")
	statusBool, err := strconv.ParseBool(status)
	if err != nil {
		s.renderer.Error(w, r, nil, model.WrapInvalid(fmt.Errorf("parse boolean with value `%s`: %w", status, err)))
		return
	}

	motionSensor, err := s.v2Service.UpdateSensor(r.Context(), id, statusBool)
	if err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("update sensor `%s`: %w", id, err))
		return
	}

	name := motionSensor.Name + " Sensor"

	stateName := "on"
	if !statusBool {
		stateName = "off"
	}

	s.renderer.Redirect(w, r, "/", renderer.NewSuccessMessage(fmt.Sprintf(updateSuccessMessage, name, stateName)))
}
