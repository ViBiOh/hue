package hue

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

// APIResponse description
type APIResponse[T any] struct {
	Data []T `json:"data"`
}

func listV2[T any](ctx context.Context, req request.Request, kind string) ([]T, error) {
	resp, err := req.Path("/clip/v2/resource"+kind).Send(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to list: %s", err)
	}

	content := APIResponse[T]{}
	if err = httpjson.Read(resp, &content); err != nil {
		return nil, fmt.Errorf("unable to parse: %s", err)
	}

	return content.Data, nil
}
