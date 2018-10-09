package worker

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

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
func (a App) Handle(ctx context.Context, p *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	if p.Action == sonos.VolumeAction {
		return a.workerVolume(ctx, p)
	}

	if p.Action == sonos.MuteAction {
		return a.workerMute(ctx, p)
	}

	return nil, fmt.Errorf(`unknown request: %s`, p)
}

func (a App) workerVolume(ctx context.Context, p *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	if parts := strings.Split(p.Payload, `|`); len(parts) == 2 {
		volume, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, errors.New(`volume is not an integer`)
		}

		if _, err := a.SetGroupVolume(ctx, parts[0], volume); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (a App) workerMute(ctx context.Context, p *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	if parts := strings.Split(p.Payload, `|`); len(parts) == 2 {
		mute, err := strconv.ParseBool(parts[1])
		if err != nil {
			return nil, errors.New(`mute is not a boolean`)
		}

		if err := a.SetGroupMute(ctx, parts[0], mute); err != nil {
			return nil, err
		}
	}

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
			return nil, fmt.Errorf(`error while listing groups: %v`, err)
		}

		household.Groups = data.Groups
		for _, group := range household.Groups {
			data, err := a.GetGroupVolume(ctx, group.ID)
			if err != nil {
				return nil, fmt.Errorf(`error while getting group volume: %v`, err)
			}

			group.Volume = data
		}
	}

	payload, err := json.Marshal(households)
	if err != nil {
		return nil, fmt.Errorf(`error while converting households payload: %v`, err)
	}

	message := provider.NewWorkerMessage(nil, sonos.Source, `households`, fmt.Sprintf(`%s`, payload))

	return []*provider.WorkerMessage{message}, nil
}
