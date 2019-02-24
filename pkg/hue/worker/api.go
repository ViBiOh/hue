package hue

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/request"
)

func hasError(content []byte) bool {
	return !bytes.Contains(content, []byte(`success`))
}

func get(ctx context.Context, url string, response interface{}) error {
	body, _, _, err := request.Get(ctx, url, nil)
	if err != nil {
		return err
	}

	content, err := request.ReadBody(body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &response); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func create(ctx context.Context, url string, payload interface{}) (*string, error) {
	body, _, _, err := request.DoJSON(ctx, url, payload, nil, http.MethodPost)
	if err != nil {
		return nil, err
	}

	content, err := request.ReadBody(body)
	if err != nil {
		return nil, err
	}

	if hasError(content) {
		return nil, errors.New(`create error: %s`, content)
	}

	var response []map[string]map[string]*string
	if err := json.Unmarshal(content, &response); err != nil {
		return nil, errors.WithStack(err)
	}

	return response[0][`success`][`id`], nil
}

func update(ctx context.Context, url string, payload interface{}) error {
	body, _, _, err := request.DoJSON(ctx, url, payload, nil, http.MethodPut)
	if err != nil {
		return err
	}

	content, err := request.ReadBody(body)
	if err != nil {
		return err
	}

	if hasError(content) {
		return errors.New(`update error: %s`, content)
	}

	return nil
}

func delete(ctx context.Context, url string) error {
	body, _, _, err := request.Do(ctx, http.MethodDelete, url, nil, nil)
	if err != nil {
		return err
	}

	content, err := request.ReadBody(body)
	if err != nil {
		return err
	}

	if hasError(content) {
		return errors.New(`delete error: %s`, content)
	}

	return nil
}
