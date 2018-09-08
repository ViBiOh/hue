package sonos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/request"
)

// ListHouseholds of user
func (a *App) ListHouseholds(ctx context.Context) ([]Household, error) {
	rawData, err := request.Get(ctx, fmt.Sprintf(`%s/households`, sonosAPIURL), http.Header{`Authorization`: []string{fmt.Sprintf(`Bearer %s`, a.accessToken)}})
	if err != nil {
		return nil, fmt.Errorf(`Error while getting households: %v`, err)
	}

	var data []Household
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling data %s: %v`, rawData, err)
	}

	return data, nil
}
