package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/iot/pkg/sonos"
)

const (
	controlURL = "https://api.ws.sonos.com/control/api/v1"
)

// HouseholdsOutput output of households endpoint
type HouseholdsOutput struct {
	Households []*sonos.Household
}

// GetHouseholds retrieves household
func (a *App) GetHouseholds(ctx context.Context) ([]*sonos.Household, error) {
	req, err := request.New().Get(fmt.Sprintf("%s/households", controlURL)).Build(ctx, nil)
	if err != nil {
		return nil, err
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, err
	}

	var data HouseholdsOutput
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
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
	req, err := request.New().Get(fmt.Sprintf("%s/households/%s/groups", controlURL, householdID)).Build(ctx, nil)
	if err != nil {
		return nil, err
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, err
	}

	var data GroupsOutput
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// GetGroupVolume retrieves volume of a Group
func (a *App) GetGroupVolume(ctx context.Context, groupID string) (*sonos.GroupVolume, error) {
	req, err := request.New().Get(fmt.Sprintf("%s/groups/%s/groupVolume", controlURL, groupID)).Build(ctx, nil)
	if err != nil {
		return nil, err
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, err
	}

	var data sonos.GroupVolume
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// SetGroupVolume defines volume of a Group
func (a *App) SetGroupVolume(ctx context.Context, groupID string, volume int) (*sonos.GroupVolume, error) {
	payload := map[string]interface{}{
		"volume": volume,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := request.New().Post(fmt.Sprintf("%s/groups/%s/groupVolume", controlURL, groupID)).Build(ctx, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, err
	}

	var data sonos.GroupVolume
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// SetGroupMute mutes volume of a Group
func (a *App) SetGroupMute(ctx context.Context, groupID string, muted bool) error {
	payload := map[string]interface{}{
		"muted": muted,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := request.New().Post(fmt.Sprintf("%s/groups/%s/groupVolume/mute", controlURL, groupID)).Build(ctx, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	_, err = a.requestWithAuth(ctx, req)
	return err
}
