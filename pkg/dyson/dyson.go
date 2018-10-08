package dyson

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/rollbar"
	"github.com/ViBiOh/httputils/pkg/tools"
)

const (
	// DysonSource constant for worker message
	DysonSource = `dyson`

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

// App stores informations
type App struct {
	account  string
	password string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	email := strings.TrimSpace(*config[`email`])
	if email == `` {
		rollbar.LogWarning(`no email provided`)
		return &App{}
	}

	password := strings.TrimSpace(*config[`password`])
	if password == `` {
		rollbar.LogWarning(`no password provided`)
		return &App{}
	}

	data := url.Values{
		`Email`:    []string{email},
		`Password`: []string{password},
	}

	loginRequest, err := http.NewRequest(http.MethodPost, fmt.Sprintf(`%s%s?country=%s`, API, authenticateEndpoint, *config[`country`]), strings.NewReader(data.Encode()))
	loginRequest.Header.Add(`Content-Type`, `application/x-www-form-urlencoded`)

	if err != nil {
		rollbar.LogError(`error while creating request to authenticate: %v`, err)
		return &App{}
	}

	payload, err := request.DoAndReadWithClient(nil, unsafeHTTPClient, loginRequest)
	if err != nil {
		rollbar.LogError(`error while authenticating: %v`, err)
		return &App{}
	}

	var authentContent map[string]string
	if err = json.Unmarshal(payload, &authentContent); err != nil {
		rollbar.LogError(`error while unmarshalling authentication content: %v`, err)
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

func (a *App) getDevices() ([]*Device, error) {
	deviceRequest, err := http.NewRequest(http.MethodGet, fmt.Sprintf(`%s%s`, API, devicesEndpoint), nil)
	if err != nil {
		return nil, fmt.Errorf(`error while creating request to list devices: %v`, err)
	}

	deviceRequest.SetBasicAuth(a.account, a.password)

	payload, err := request.DoAndReadWithClient(nil, unsafeHTTPClient, deviceRequest)
	if err != nil {
		return nil, fmt.Errorf(`error while listing devices: %v`, err)
	}

	var devices []*Device
	if err = json.Unmarshal(payload, &devices); err != nil {
		return nil, fmt.Errorf(`error while unmarshalling devices content: %v`, err)
	}

	return devices, nil
}

func (a *App) isReady() bool {
	return a.account != `` && a.password != ``
}

// GetData return data for Dashboard rendering
func (a *App) GetData() interface{} {
	if !a.isReady() {
		return nil
	}

	devices, err := a.getDevices()
	if err != nil {
		rollbar.LogError(`error while getting devices: %v`, err)
	}

	return devices
}

// Handler for request. Should be use with net/http
func (a App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if err := httpjson.ResponseJSON(w, http.StatusOK, a.GetData(), httpjson.IsPretty(r)); err != nil {
				httperror.InternalServerError(w, err)
			}
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}
