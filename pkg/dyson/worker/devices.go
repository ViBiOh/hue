package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/iot/pkg/dyson"
)

func (a *App) getDevices(ctx context.Context) ([]*dyson.Device, error) {
	deviceRequest, err := http.NewRequest(http.MethodGet, fmt.Sprintf(`%s%s`, API, devicesEndpoint), nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	deviceRequest.SetBasicAuth(a.account, a.password)

	body, _, _, err := request.DoAndReadWithClient(ctx, unsafeHTTPClient, deviceRequest)
	if err != nil {
		return nil, err
	}

	payload, err := request.ReadBody(body)
	if err != nil {
		return nil, err
	}

	var devices []*dyson.Device
	if err = json.Unmarshal(payload, &devices); err != nil {
		return nil, errors.WithStack(err)
	}

	for _, device := range devices {
		credentials, err := decipherLocalCredentials(device.LocalCredentials)
		if err != nil {
			return nil, err
		}

		device.Credentials = credentials
	}

	return devices, nil
}
