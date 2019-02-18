package worker

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"sync"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/netatmo"
	"github.com/ViBiOh/iot/pkg/provider"
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
		accessToken:  fs.String(tools.ToCamel(fmt.Sprintf(`%sAccessToken`, prefix)), ``, `[netatmo] Access Token`),
		refreshToken: fs.String(tools.ToCamel(fmt.Sprintf(`%sRefreshToken`, prefix)), ``, `[netatmo] Refresh Token`),
		clientID:     fs.String(tools.ToCamel(fmt.Sprintf(`%sClientID`, prefix)), ``, `[netatmo] Client ID`),
		clientSecret: fs.String(tools.ToCamel(fmt.Sprintf(`%sClientSecret`, prefix)), ``, `[netatmo] Client Secret`),
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
	return netatmo.Source
}

// Handle handle worker requests for Netatmo
func (a *App) Handle(ctx context.Context, message *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	return nil, nil
}

// Ping send to worker updated data
func (a *App) Ping(ctx context.Context) ([]*provider.WorkerMessage, error) {
	stationsData, err := a.getStationsData(ctx, true)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(stationsData.Body.Devices)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	message := provider.NewWorkerMessage(nil, netatmo.Source, netatmo.DevicesAction, fmt.Sprintf(`%s`, payload))

	return []*provider.WorkerMessage{message}, nil
}
