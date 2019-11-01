package hue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/v3/pkg/request"
)

func hasError(content []byte) bool {
	return !bytes.Contains(content, []byte("success"))
}

func get(ctx context.Context, url string, response interface{}) error {
	resp, err := request.Get(ctx, url, nil)
	if err != nil {
		return err
	}

	content, err := request.ReadBodyResponse(resp)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &response); err != nil {
		return err
	}

	return nil
}

func create(ctx context.Context, url string, payload interface{}) (*string, error) {
	resp, err := request.PostJSON(ctx, url, payload, nil)
	if err != nil {
		return nil, err
	}

	content, err := request.ReadBodyResponse(resp)
	if err != nil {
		return nil, err
	}

	if hasError(content) {
		return nil, fmt.Errorf("create error: %s", content)
	}

	var response []map[string]map[string]*string
	if err := json.Unmarshal(content, &response); err != nil {
		return nil, err
	}

	return response[0]["success"]["id"], nil
}

func update(ctx context.Context, url string, payload interface{}) error {
	req, err := request.JSON(ctx, http.MethodPut, url, payload, nil)
	if err != nil {
		return err
	}

	resp, err := request.Do(ctx, req)
	if err != nil {
		return err
	}

	content, err := request.ReadBodyResponse(resp)
	if err != nil {
		return err
	}

	if hasError(content) {
		return fmt.Errorf("update error: %s", content)
	}

	return nil
}

func delete(ctx context.Context, url string) error {
	req, err := request.New(ctx, http.MethodDelete, url, nil, nil)
	if err != nil {
		return err
	}

	resp, err := request.Do(ctx, req)
	if err != nil {
		return err
	}

	content, err := request.ReadBodyResponse(resp)
	if err != nil {
		return err
	}

	if hasError(content) {
		return fmt.Errorf("delete error: %s", content)
	}

	return nil
}
