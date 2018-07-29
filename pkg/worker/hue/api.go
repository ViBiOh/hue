package hue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/request"
)

func hasError(content []byte) bool {
	return !bytes.Contains(content, []byte(`success`))
}

func get(ctx context.Context, url string, response interface{}) error {
	content, err := request.Get(ctx, url, nil)

	if err != nil {
		return fmt.Errorf(`Error while sending get request: %v`, err)
	}

	if err := json.Unmarshal(content, &response); err != nil {
		return fmt.Errorf(`Error while parsing response %s: %v`, content, err)
	}

	return nil
}

func create(ctx context.Context, url string, payload interface{}) (*string, error) {
	content, err := request.DoJSON(ctx, url, payload, nil, http.MethodPost)

	if err != nil {
		return nil, fmt.Errorf(`Error while sending post request: %v`, err)
	}

	if hasError(content) {
		return nil, fmt.Errorf(`Error while sending post request: %s`, content)
	}

	var response []map[string]map[string]*string
	if err := json.Unmarshal(content, &response); err != nil {
		return nil, fmt.Errorf(`Error while parsing response %s: %v`, content, err)
	}

	return response[0][`success`][`id`], nil
}

func update(ctx context.Context, url string, payload interface{}) error {
	content, err := request.DoJSON(ctx, url, payload, nil, http.MethodPut)

	if err != nil {
		return fmt.Errorf(`Error while sending put request: %v`, err)
	}

	if hasError(content) {
		return fmt.Errorf(`Error while sending put request: %s`, content)
	}

	return nil
}

func delete(ctx context.Context, url string) error {
	content, err := request.Do(ctx, url, nil, nil, http.MethodDelete)

	if err != nil {
		return fmt.Errorf(`Error while sending delete request: %v`, err)
	}

	if hasError(content) {
		return fmt.Errorf(`Error while sending delete request: %s`, content)
	}

	return nil
}
