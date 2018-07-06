package dyson

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	// DysonSource constant for worker message
	DysonSource = `dyson`

	// API of Dyson Link
	API = `https://api.cp.dyson.com`
)

var unsafeHTTPClient = http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// App stores informations
type App struct {
	account  string
	password string
	hub      provider.Hub
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	email := strings.TrimSpace(*config[`email`])
	if email == `` {
		log.Print(`[dyson] No email provided`)
		return &App{}
	}

	password := strings.TrimSpace(*config[`password`])
	if password == `` {
		log.Print(`[dyson] No password provided`)
		return &App{}
	}

	data := url.Values{
		`Email`:    []string{email},
		`Password`: []string{password},
	}

	loginRequest, err := http.NewRequest(http.MethodPost, fmt.Sprintf(`%s/v1/userregistration/authenticate?country=%s`, API, *config[`country`]), strings.NewReader(data.Encode()))
	loginRequest.Header.Add(`Content-Type`, `application/x-www-form-urlencoded`)

	if err != nil {
		log.Printf(`[dyson] Error while creating request to authenticate to Dyson: %v`, err)
		return &App{}
	}

	payload, err := request.DoAndReadWithClient(nil, unsafeHTTPClient, loginRequest)
	if err != nil {
		log.Printf(`[dyson] Error while authenticating to Dyson: %v`, err)
		return &App{}
	}

	var authentContent map[string]string
	if err = json.Unmarshal(payload, &authentContent); err != nil {
		log.Printf(`[dyson] Error while unmarshalling authentication content: %v`, err)
		return &App{}
	}

	return &App{
		account:  authentContent[`Account`],
		password: authentContent[`Password`],
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`email`:    flag.String(tools.ToCamel(fmt.Sprintf(`%sEmail`, prefix)), ``, `Dyson Link Email`),
		`password`: flag.String(tools.ToCamel(fmt.Sprintf(`%sPassword`, prefix)), ``, `Dyson Link Password`),
		`country`:  flag.String(tools.ToCamel(fmt.Sprintf(`%sCountry`, prefix)), `FR`, `Dyson Link Country`),
	}
}

// SetHub receive Hub during init of it
func (a *App) SetHub(hub provider.Hub) {
	a.hub = hub
}

// GetWorkerSource get source of message in websocket
func (a *App) GetWorkerSource() string {
	return DysonSource
}

// GetData return data for Dashboard rendering
func (a *App) GetData(ctx context.Context) interface{} {
	return nil
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(message *provider.WorkerMessage) error {
	return fmt.Errorf(`Unknown worker command: %s`, message.Type)
}
