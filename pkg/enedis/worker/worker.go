package worker

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/enedis"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	loginURL   = `https://espace-client-connexion.enedis.fr/auth/UI/Login`
	consumeURL = `https://espace-client-particuliers.enedis.fr/group/espace-particuliers/suivi-de-consommation?`

	frenchDateFormat = `02/01/2006`
)

// Config of package
type Config struct {
	email    *string
	password *string
}

// App of package
type App struct {
	email    string
	password string
	cookie   string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		email:    fs.String(tools.ToCamel(fmt.Sprintf(`%sEmail`, prefix)), ``, `[enedis] Email`),
		password: fs.String(tools.ToCamel(fmt.Sprintf(`%sPassword`, prefix)), ``, `[enedis] Password`),
	}
}

// New creates new App from Config
func New(config Config) *App {
	return &App{
		email:    strings.TrimSpace(*config.email),
		password: strings.TrimSpace(*config.password),
	}
}

// GetSource returns source name
func (a *App) GetSource() string {
	return enedis.Source
}

// Handle handle worker requests for Enedis
func (a *App) Handle(context.Context, *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	return nil, nil
}

// Start the package
func (a *App) Start() {
	if !a.Enabled() {
		logger.Warn(`no config provided`)
		return
	}

	if err := a.Login(); err != nil {
		logger.Error(`%+v`, err)
	}
}

// Ping send to worker updated data
func (a *App) Ping(ctx context.Context) ([]*provider.WorkerMessage, error) {
	_, err := a.GetData(ctx, true)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Enabled checks if worker is enabled
func (a *App) Enabled() bool {
	return a.email != `` && a.password != ``
}
