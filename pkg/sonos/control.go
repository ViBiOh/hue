package sonos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	controlURL = `https://api.ws.sonos.com/control/api/v1`
)

// GroupsOutput output of groups endpoint
type GroupsOutput struct {
	Players []Player
	Groups  []Group
}

// GetHouseholds of user
func (a *App) GetHouseholds(ctx context.Context) ([]Household, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(`%s/households`, controlURL), nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting households: %v`, err)
	}

	var data map[string][]Household
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling data %s: %v`, rawData, err)
	}

	return data[`households`], nil
}

// GetGroups of household
func (a *App) GetGroups(ctx context.Context, householdID string) (*GroupsOutput, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(`%s/households/%s/groups`, controlURL, householdID), nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	rawData, err := a.requestWithAuth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting groups: %v`, err)
	}

	var data GroupsOutput
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling data %s: %v`, rawData, err)
	}

	return &data, nil
}
