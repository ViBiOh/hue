package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/request"
)

func hasError(content []byte) bool {
	return !bytes.Contains(content, []byte(`success`))
}

func get(url string, response interface{}) error {
	content, err := request.Get(url, nil)
	if err != nil {
		return fmt.Errorf(`Error while sending get request: %v`, err)
	}

	if hasError(content) {
		return fmt.Errorf(`Error while sending get request: %s`, content)
	}

	if err := json.Unmarshal(content, &response); err != nil {
		return fmt.Errorf(`Error while parsing response: %v`, err)
	}

	return nil
}

func create(url string, payload interface{}) (*string, error) {
	content, err := request.DoJSON(url, payload, nil, http.MethodPost)
	if err != nil {
		return nil, fmt.Errorf(`Error while sending post request: %v`, err)
	}

	if hasError(content) {
		return nil, fmt.Errorf(`Error while sending post request: %s`, content)
	}

	var response []map[string]map[string]*string
	if err := json.Unmarshal(content, &response); err != nil {
		return nil, fmt.Errorf(`Error while parsing response: %s`, err)
	}

	return response[0][`success`][`id`], nil
}

func update(url string, payload interface{}) error {
	content, err := request.DoJSON(url, payload, nil, http.MethodPut)
	if err != nil {
		return fmt.Errorf(`Error while sending put request: %v`, err)
	}

	if hasError(content) {
		return fmt.Errorf(`Error while sending put request: %s`, content)
	}

	return nil
}

func delete(url string) error {
	content, err := request.Do(url, nil, nil, http.MethodDelete)
	if err != nil {
		return fmt.Errorf(`Error while sending delete request: %v`, err)
	}

	if hasError(content) {
		return fmt.Errorf(`Error while sending delete request: %s`, content)
	}

	return nil
}
