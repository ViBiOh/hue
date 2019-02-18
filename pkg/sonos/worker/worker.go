package worker

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/provider"
	"github.com/ViBiOh/iot/pkg/sonos"
)

// Config of package
type Config struct {
	accessToken  *string
	refreshToken *string
	clientID     *string
	clientSecret *string
}

// App of package
type App struct {
	clientID     string
	clientSecret string
	accessToken  string
	refreshToken string
	mutex        sync.RWMutex
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		accessToken:  fs.String(tools.ToCamel(fmt.Sprintf(`%sAccessToken`, prefix)), ``, `[sonos] Access Token`),
		refreshToken: fs.String(tools.ToCamel(fmt.Sprintf(`%sRefreshToken`, prefix)), ``, `[sonos] Refresh Token`),
		clientID:     fs.String(tools.ToCamel(fmt.Sprintf(`%sClientID`, prefix)), ``, `[sonos] Client ID`),
		clientSecret: fs.String(tools.ToCamel(fmt.Sprintf(`%sClientSecret`, prefix)), ``, `[sonos] Client Secret`),
	}
}

// New creates new App from Config
func New(config Config) *App {
	return &App{
		clientID:     *config.clientID,
		clientSecret: *config.clientSecret,
		accessToken:  *config.accessToken,
		refreshToken: *config.refreshToken,
	}
}

// Enabled checks if worker is enabled
func (a *App) Enabled() bool {
	return a.clientID != `` && a.clientSecret != `` && a.accessToken != `` && a.refreshToken != ``
}

// GetSource returns source name
func (a *App) GetSource() string {
	return sonos.Source
}

// Handle handle worker requests for Netatmo
func (a *App) Handle(ctx context.Context, p *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	if p.Action == sonos.VolumeAction {
		return a.workerVolume(ctx, p)
	}

	if p.Action == sonos.MuteAction {
		return a.workerMute(ctx, p)
	}

	return nil, errors.New(`unknown request: %s`, p)
}

func (a *App) workerVolume(ctx context.Context, p *provider.WorkerMessage) (*provider.WorkerMessage, error) {
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

func (a *App) workerMute(ctx context.Context, p *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	if parts := strings.Split(p.Payload, `|`); len(parts) == 2 {
		mute, err := strconv.ParseBool(parts[1])
		if err != nil {
			return nil, errors.New(`mute is not a boolean`)
		}

		if err := a.SetGroupMute(ctx, parts[0], mute); err != nil {
			return nil, err
		}

		return provider.NewWorkerMessage(p, sonos.Source, "mute", fmt.Sprintf(`%s|%t`, parts[0], mute)), nil
	}

	return nil, nil
}

// Ping send to worker updated data
func (a *App) Ping(ctx context.Context) ([]*provider.WorkerMessage, error) {
	households, err := a.GetHouseholds(ctx)
	if err != nil {
		return nil, err
	}

	for _, household := range households {
		data, err := a.GetGroups(ctx, household.ID)
		if err != nil {
			return nil, err
		}

		household.Groups = data.Groups
		for _, group := range household.Groups {
			data, err := a.GetGroupVolume(ctx, group.ID)
			if err != nil {
				return nil, err
			}

			group.Volume = data
		}
	}

	payload, err := json.Marshal(households)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	message := provider.NewWorkerMessage(nil, sonos.Source, `households`, fmt.Sprintf(`%s`, payload))

	return []*provider.WorkerMessage{message}, nil
}
