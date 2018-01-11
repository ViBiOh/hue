package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/iot/hue"
)

func (a *App) listRules() (map[string]*hue.Rule, error) {
	content, err := httputils.GetRequest(fmt.Sprintf(`%s/rules`, a.bridgeURL), nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while sending get request: %v`, err)
	}

	var response map[string]*hue.Rule
	if err := json.Unmarshal(content, &response); err != nil {
		return nil, fmt.Errorf(`Error while parsing response: %v`, err)
	}

	return response, nil
}

func (a *App) createRule(o *hue.Rule) error {
	content, err := httputils.RequestJSON(fmt.Sprintf(`%s/rules`, a.bridgeURL), o, nil, http.MethodPost)
	if err != nil {
		return fmt.Errorf(`Error while sending post request: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while sending post request: %s`, content)
	}

	var response []map[string]map[string]string
	if err := json.Unmarshal(content, &response); err != nil {
		return fmt.Errorf(`Error while parsing result: %s`, err)
	}

	o.ID = response[0][`success`][`id`]

	return nil
}

func (a *App) updateRule(o *hue.Rule) error {
	content, err := httputils.RequestJSON(fmt.Sprintf(`%s/rules/%s`, a.bridgeURL, o.ID), o, nil, http.MethodPut)
	if err != nil {
		return fmt.Errorf(`Error while sending put request: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while sending put request: %s`, content)
	}

	return nil
}

func (a *App) deleteRule(id string) error {
	content, err := httputils.Request(fmt.Sprintf(`%s/rules/%s`, a.bridgeURL, id), nil, nil, http.MethodDelete)
	if err != nil {
		return fmt.Errorf(`Error while sending delete request: %v`, err)
	}
	if !bytes.Contains(content, []byte(`success`)) {
		return fmt.Errorf(`Error while sending delete request: %s`, content)
	}

	return nil
}

func (a *App) cleanRules() error {
	rules, err := a.listRules()
	if err != nil {
		return fmt.Errorf(`Error while listing rules: %v`, err)
	}

	for key := range rules {
		if err := a.deleteRule(key); err != nil {
			return fmt.Errorf(`Error while deleting rule: %v`, err)
		}
	}

	return nil
}
