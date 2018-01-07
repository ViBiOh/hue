package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils"
)

func (a *App) listSchedules() (map[string]interface{}, error) {
	content, err := httputils.GetRequest(a.bridgeURL+`/schedules`, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting schedules: %v`, err)
	}

	var schedules map[string]interface{}
	if err := json.Unmarshal(content, &schedules); err != nil {
		return nil, fmt.Errorf(`Error while parsing schedules: %v`, err)
	}

	return schedules, nil
}

func (a *App) deleteSchedule(id string) error {
	content, err := httputils.Request(a.bridgeURL+`/schedules/`+id, nil, nil, http.MethodDelete)
	if err != nil {
		return fmt.Errorf(`Error while deleting schedule: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while deleting schedule: %s`, content)
	}

	return nil
}

func (a *App) cleanSchedules() error {
	schedules, err := a.listSchedules()
	if err != nil {
		return fmt.Errorf(`Error while listing schedules: %v`, err)
	}

	for key := range schedules {
		if err := a.deleteSchedule(key); err != nil {
			return fmt.Errorf(`Error while deleting schedule: %v`, err)
		}
	}

	return nil
}
