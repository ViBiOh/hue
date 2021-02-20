package hue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func hasError(content []byte) bool {
	return !bytes.Contains(content, []byte("success"))
}

func get(ctx context.Context, url string, response interface{}) error {
	resp, err := request.New().Get(url).Send(ctx, nil)
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

func create(ctx context.Context, url string, payload interface{}) (string, error) {
	resp, err := request.New().Post(url).JSON(ctx, payload)
	if err != nil {
		return "", err
	}

	content, err := request.ReadBodyResponse(resp)
	if err != nil {
		return "", err
	}

	if hasError(content) {
		return "", fmt.Errorf("create error: %s", content)
	}

	var response []map[string]map[string]string
	if err := json.Unmarshal(content, &response); err != nil {
		return "", err
	}

	return response[0]["success"]["id"], nil
}

func update(ctx context.Context, url string, payload interface{}) error {
	resp, err := request.New().Put(url).JSON(ctx, payload)
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

func remove(ctx context.Context, url string) error {
	resp, err := request.New().Delete(url).Send(ctx, nil)
	if err != nil {
		return err
	}

	content, err := request.ReadBodyResponse(resp)
	if err != nil {
		return err
	}

	if hasError(content) {
		return fmt.Errorf("remove error: %s", content)
	}

	return nil
}
