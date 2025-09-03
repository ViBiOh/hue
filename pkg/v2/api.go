package v2

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

// APIResponse description
type APIResponse[T any] struct {
	Data   []T `json:"data"`
	Errors []struct {
		Description string `json:"description"`
	} `json:"errors"`
}

func list[T any](ctx context.Context, req request.Request, kind string) (output []T, err error) {
	resp, err := req.Path(path.Join("/clip/v2/resource", kind)).Send(ctx, nil)
	if err != nil {
		return output, fmt.Errorf("list: %w", err)
	}

	content, err := httpjson.Read[APIResponse[T]](resp)
	if err != nil {
		return output, fmt.Errorf("parse: %w", err)
	}

	output = content.Data

	return output, err
}

func stream[T any](ctx context.Context, req request.Request, kind string, output chan<- T) (err error) {
	resp, err := req.Path(path.Join("/clip/v2/resource", kind)).Send(ctx, nil)
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}

	return httpjson.Stream(resp.Body, output, "data", true)
}

func get[T any](ctx context.Context, req request.Request, kind, id string) (output T, err error) {
	resp, err := req.Path(path.Join("/clip/v2/resource", kind, id)).Send(ctx, nil)
	if err != nil {
		return output, fmt.Errorf("get: %w", err)
	}

	content, err := httpjson.Read[APIResponse[T]](resp)
	if err != nil {
		return output, fmt.Errorf("parse: %w", err)
	}

	if len(content.Data) == 0 {
		return output, errors.New("not found")
	}

	return content.Data[0], nil
}
