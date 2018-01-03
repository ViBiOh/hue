package wemo

import (
	"flag"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/iot"
)

// App stores informations and secret of API
type App struct {
	webHookKey string
	iotApp     *iot.App
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string, iotApp *iot.App) *App {
	return &App{
		webHookKey: *config[`webhookKey`],
		iotApp:     iotApp,
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`webhookKey`: flag.String(tools.ToCamel(prefix+`WebHook`), ``, `[wemo] WebHook Key`),
	}
}

// Handler create Handler from Flags' config
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rawData []byte
		var err error

		if r.URL.Path == `/on` {
			rawData, err = httputils.GetBody(`https://maker.ifttt.com/trigger/wemo_plug_on/with/key/`+a.webHookKey, nil)
		} else if r.URL.Path == `/off` {
			rawData, err = httputils.GetBody(`https://maker.ifttt.com/trigger/wemo_plug_off/with/key/`+a.webHookKey, nil)
		} else {
			a.iotApp.RenderDashboard(w, r, http.StatusNotFound, &iot.Message{Level: `error`, Content: `Unknown command`})
			return
		}

		if err != nil || strings.HasPrefix(string(rawData), `<!DOCTYPE`) {
			a.iotApp.RenderDashboard(w, r, http.StatusInternalServerError, &iot.Message{Level: `error`, Content: `Error while requesting WeMo`})
		} else {
			a.iotApp.RenderDashboard(w, r, http.StatusOK, nil)
		}
	})
}
