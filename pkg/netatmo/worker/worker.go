package worker

import (
	"context"
	"flag"
	"fmt"

	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/netatmo"
	"github.com/ViBiOh/iot/pkg/provider"
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
		`accessToken`:  flag.String(tools.ToCamel(fmt.Sprintf(`%sAccessToken`, prefix)), ``, `[netatmo] Access Token`),
		`refreshToken`: flag.String(tools.ToCamel(fmt.Sprintf(`%sRefreshToken`, prefix)), ``, `[netatmo] Refresh Token`),
		`clientID`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sClientID`, prefix)), ``, `[netatmo] Client ID`),
		`clientSecret`: flag.String(tools.ToCamel(fmt.Sprintf(`%sClientSecret`, prefix)), ``, `[netatmo] Client Secret`),
	}
}

// GetSource returns source name for WS calls
func (a App) GetSource() string {
	return netatmo.Source
}

// Handle handle worker requests for Netatmo
func (a App) Handle(ctx context.Context, message *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	return nil, nil
}

// Ping send to worker updated data
func (a App) Ping(ctx context.Context) ([]*provider.WorkerMessage, error) {
	stationsData, err := a.getStationsData(ctx, true)
	if err != nil {
		return nil, err
	}

	message := provider.NewWorkerMessage(``, netatmo.Source, `devices`, stationsData.Body.Devices)

	return []*provider.WorkerMessage{message}, nil
}
