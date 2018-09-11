package sonos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/request"
)

const (
	controlURL = `https://api.ws.sonos.com/control/api/v1`
)

// HouseholdsOutput output of households endpoint
type HouseholdsOutput struct {
	Households []*Household
}

// GetHouseholds retrieves household
func (a *App) GetHouseholds(ctx context.Context) ([]*Household, error) {
	req, err := request.New(http.MethodGet, fmt.Sprintf(`%s/households`, controlURL), nil, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting households: %v - %s`, err, rawData)
	}

	var data HouseholdsOutput
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling data %s: %v`, rawData, err)
	}

	return data.Households, nil
}

// GroupsOutput output of groups endpoint
type GroupsOutput struct {
	Players []*Player
	Groups  []*Group
}

// GetGroups retrieves groups of a Household
func (a *App) GetGroups(ctx context.Context, householdID string) (*GroupsOutput, error) {
	req, err := request.New(http.MethodGet, fmt.Sprintf(`%s/households/%s/groups`, controlURL, householdID), nil, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting groups: %v - %s`, err, rawData)
	}

	var data GroupsOutput
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling data %s: %v`, rawData, err)
	}

	return &data, nil
}

// GetGroupVolume retrieves volume of a Group
func (a *App) GetGroupVolume(ctx context.Context, groupID string) (*GroupVolume, error) {
	req, err := request.New(http.MethodGet, fmt.Sprintf(`%s/groups/%s/groupVolume`, controlURL, groupID), nil, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting group volume: %v - %s`, err, rawData)
	}

	var data GroupVolume
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling data %s: %v`, rawData, err)
	}

	return &data, nil
}

// SetGroupVolume retrieves volume of a Group
func (a *App) SetGroupVolume(ctx context.Context, groupID string, volume int) (*GroupVolume, error) {
	payload := GroupVolume{
		Volume: volume,
	}

	req, err := request.JSON(http.MethodPost, fmt.Sprintf(`%s/groups/%s/groupVolume`, controlURL, groupID), payload, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf(`Error while setting group volume: %v - %s`, err, rawData)
	}

	var data GroupVolume
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling data %s: %v`, rawData, err)
	}

	return &data, nil
}
