package worker

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/dyson"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	// API of Dyson Link
	API = `https://api.cp.dyson.com`

	authenticateEndpoint = `/v1/userregistration/authenticate`
	devicesEndpoint      = `/v1/provisioningservice/manifest`
)

var unsafeHTTPClient = http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// Config of package
type Config struct {
	email    *string
	password *string
	country  *string
}

// App of package
type App struct {
	account  string
	password string
	devices  []*dyson.Device
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		email:    fs.String(tools.ToCamel(fmt.Sprintf(`%sEmail`, prefix)), ``, `Dyson Link Email`),
		password: fs.String(tools.ToCamel(fmt.Sprintf(`%sPassword`, prefix)), ``, `Dyson Link Password`),
		country:  fs.String(tools.ToCamel(fmt.Sprintf(`%sCountry`, prefix)), `FR`, `Dyson Link Country`),
	}
}

// New creates new App from Config
func New(config Config) *App {
	email := strings.TrimSpace(*config.email)
	if email == `` {
		logger.Warn(`no email provided`)
		return &App{}
	}

	password := strings.TrimSpace(*config.password)
	if password == `` {
		logger.Warn(`no password provided`)
		return &App{}
	}

	authContent, err := getAuth(email, password, *config.country)
	if err != nil {
		logger.Error(`%+v`, err)
		return &App{}
	}

	app := &App{
		account:  authContent[`Account`],
		password: authContent[`Password`],
	}

	devices, err := app.getDevices(nil)
	if err != nil {
		logger.Error(`%+v`, err)
	} else {
		app.devices = devices
	}

	return app
}

// GetSource returns source name
func (a *App) GetSource() string {
	return dyson.Source
}

// Handle handle worker requests for Netatmo
func (a *App) Handle(ctx context.Context, p *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	return nil, nil
}

// Ping send to worker updated data
func (a *App) Ping(ctx context.Context) ([]*provider.WorkerMessage, error) {
	return nil, nil
}
