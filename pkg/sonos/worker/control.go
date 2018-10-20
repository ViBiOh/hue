package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/iot/pkg/sonos"
)

const (
	controlURL = `https://api.ws.sonos.com/control/api/v1`
)

// HouseholdsOutput output of households endpoint
type HouseholdsOutput struct {
	Households []*sonos.Household
}

// GetHouseholds retrieves household
func (a *App) GetHouseholds(ctx context.Context) ([]*sonos.Household, error) {
	req, err := request.New(http.MethodGet, fmt.Sprintf(`%s/households`, controlURL), nil, nil)
	if err != nil {
		return nil, err
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, err
	}

	var data HouseholdsOutput
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, errors.WithStack(err)
	}

	return data.Households, nil
}

// GroupsOutput output of groups endpoint
type GroupsOutput struct {
	Players []*sonos.Player
	Groups  []*sonos.Group
}

// GetGroups retrieves groups of a Household
func (a *App) GetGroups(ctx context.Context, householdID string) (*GroupsOutput, error) {
	req, err := request.New(http.MethodGet, fmt.Sprintf(`%s/households/%s/groups`, controlURL, householdID), nil, nil)
	if err != nil {
		return nil, err
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, err
	}

	var data GroupsOutput
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, errors.WithStack(err)
	}

	return &data, nil
}

// GetGroupVolume retrieves volume of a Group
func (a *App) GetGroupVolume(ctx context.Context, groupID string) (*sonos.GroupVolume, error) {
	req, err := request.New(http.MethodGet, fmt.Sprintf(`%s/groups/%s/groupVolume`, controlURL, groupID), nil, nil)
	if err != nil {
		return nil, err
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, err
	}

	var data sonos.GroupVolume
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, errors.WithStack(err)
	}

	return &data, nil
}

// SetGroupVolume defines volume of a Group
func (a *App) SetGroupVolume(ctx context.Context, groupID string, volume int) (*sonos.GroupVolume, error) {
	payload := map[string]interface{}{
		`volume`: volume,
	}

	req, err := request.JSON(http.MethodPost, fmt.Sprintf(`%s/groups/%s/groupVolume`, controlURL, groupID), payload, nil)
	if err != nil {
		return nil, err
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, err
	}

	var data sonos.GroupVolume
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, errors.WithStack(err)
	}

	return &data, nil
}

// SetGroupMute mutes volume of a Group
func (a *App) SetGroupMute(ctx context.Context, groupID string, muted bool) error {
	payload := map[string]interface{}{
		`muted`: muted,
	}

	req, err := request.JSON(http.MethodPost, fmt.Sprintf(`%s/groups/%s/mute`, controlURL, groupID), payload, nil)
	if err != nil {
		return err
	}

	_, err = a.requestWithAuth(ctx, req)
	return err
}
