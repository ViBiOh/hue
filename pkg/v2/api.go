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
		err = fmt.Errorf("list: %w", err)
		return
	}

	content, err := httpjson.Read[APIResponse[T]](resp)
	if err != nil {
		err = fmt.Errorf("parse: %w", err)
		return
	}

	output = content.Data

	return
}

func stream[T any](ctx context.Context, req request.Request, kind string, output chan<- T) (err error) {
	resp, err := req.Path(path.Join("/clip/v2/resource", kind)).Send(ctx, nil)
	if err != nil {
		err = fmt.Errorf("list: %w", err)
		return
	}

	return httpjson.Stream[T](resp.Body, output, "data", true)
}

func get[T any](ctx context.Context, req request.Request, kind, id string) (output T, err error) {
	resp, err := req.Path(path.Join("/clip/v2/resource", kind, id)).Send(ctx, nil)
	if err != nil {
		err = fmt.Errorf("get: %w", err)
		return
	}

	content, err := httpjson.Read[APIResponse[T]](resp)
	if err != nil {
		err = fmt.Errorf("parse: %w", err)
		return
	}

	if len(content.Data) == 0 {
		err = errors.New("not found")
		return
	}

	return content.Data[0], nil
}
