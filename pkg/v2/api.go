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
		err = fmt.Errorf("unable to list: %s", err)
		return
	}

	content := APIResponse[T]{}
	if err = httpjson.Read(resp, &content); err != nil {
		err = fmt.Errorf("unable to parse: %s", err)
		return
	}

	output = content.Data

	return
}

func get[T any](ctx context.Context, req request.Request, kind, id string) (output T, err error) {
	resp, err := req.Path(path.Join("/clip/v2/resource", kind, id)).Send(ctx, nil)
	if err != nil {
		err = fmt.Errorf("unable to get: %s", err)
		return
	}

	content := APIResponse[T]{}
	if err = httpjson.Read(resp, &content); err != nil {
		err = fmt.Errorf("unable to parse: %s", err)
		return
	}

	if len(content.Data) == 0 {
		err = errors.New("not found")
		return
	}

	return content.Data[0], nil
}
