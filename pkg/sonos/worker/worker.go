package worker

import (
	"context"
	"flag"
	"fmt"

	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/provider"
	"github.com/ViBiOh/iot/pkg/sonos"
)

// App stores informations
type App struct {
	clientID     string
	clientSecret string
	accessToken  string
	refreshToken string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	return &App{
		clientID:     *config[`clientID`],
		clientSecret: *config[`clientSecret`],
		accessToken:  *config[`accessToken`],
		refreshToken: *config[`refreshToken`],
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`accessToken`:  flag.String(tools.ToCamel(fmt.Sprintf(`%sAccessToken`, prefix)), ``, `[sonos] Access Token`),
		`refreshToken`: flag.String(tools.ToCamel(fmt.Sprintf(`%sRefreshToken`, prefix)), ``, `[sonos] Refresh Token`),
		`clientID`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sClientID`, prefix)), ``, `[sonos] Client ID`),
		`clientSecret`: flag.String(tools.ToCamel(fmt.Sprintf(`%sClientSecret`, prefix)), ``, `[sonos] Client Secret`),
	}
}

// GetSource returns source name for WS calls
func (a App) GetSource() string {
	return sonos.Source
}

// Handle handle worker requests for Netatmo
func (a App) Handle(ctx context.Context, message *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	return nil, nil
}

// Ping send to worker updated data
func (a App) Ping(ctx context.Context) ([]*provider.WorkerMessage, error) {
	households, err := a.GetHouseholds(ctx)
	if err != nil {
		return nil, fmt.Errorf(`error while listing households: %v`, err)
	}

	for _, household := range households {
		data, err := a.GetGroups(ctx, household.ID)
		if err != nil {
			return nil, fmt.Errorf(`[sonos] Error while listing groups: %v`, err)
		}

		household.Groups = data.Groups
		for _, group := range household.Groups {
			data, err := a.GetGroupVolume(ctx, group.ID)
			if err != nil {
				return nil, fmt.Errorf(`[sonos] Error while getting group volume: %v`, err)
			}

			group.Volume = data
		}
	}

	message := provider.NewWorkerMessage(``, sonos.Source, `households`, households)

	return []*provider.WorkerMessage{message}, nil
}
