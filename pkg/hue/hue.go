package hue

import (
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	v2 "github.com/ViBiOh/hue/pkg/v2"
)

// App stores information and secret of API
type App struct {
	apiHandler     http.Handler
	v2App          *v2.App
	scenes         map[string]Scene
	lights         map[string]Light
	groups         map[string]Group
	schedules      map[string]Schedule
	bridgeUsername string
	bridgeURL      string
	configFileName string
	rendererApp    renderer.App
	syncers        []syncer
	mutex          sync.RWMutex
	update         bool
}

// Config of package
type Config struct {
	bridgeIP       *string
	bridgeUsername *string
	config         *string
	update         *bool
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		bridgeIP:       flags.New("BridgeIP", "IP of Bridge").Prefix(prefix).DocPrefix("hue").String(fs, "", nil),
		bridgeUsername: flags.New("Username", "Username for Bridge").Prefix(prefix).DocPrefix("hue").String(fs, "", nil),
		config:         flags.New("Config", "Configuration filename").Prefix(prefix).DocPrefix("hue").String(fs, "", nil),
		update:         flags.New("Update", "Update configuration from file").Prefix(prefix).DocPrefix("hue").Bool(fs, false, nil),
	}
}

// New creates new App from Config
func New(config Config, renderer renderer.App, v2App *v2.App) (*App, error) {
	bridgeAddress := strings.TrimSpace(*config.bridgeIP)
	bridgeUsername := strings.TrimSpace(*config.bridgeUsername)

	app := App{
		bridgeURL:      fmt.Sprintf("http://%s/api/%s", bridgeAddress, bridgeUsername),
		bridgeUsername: bridgeUsername,
		configFileName: strings.TrimSpace(*config.config),
		update:         *config.update,
		rendererApp:    renderer,
		v2App:          v2App,
	}

	app.syncers = []syncer{
		app.syncGroups,
		app.syncSchedules,
		app.syncScenes,
	}

	app.apiHandler = http.StripPrefix(apiPath, app.Handler())

	return &app, nil
}

// TemplateFunc for rendering GUI
func (a *App) TemplateFunc(w http.ResponseWriter, r *http.Request) (renderer.Page, error) {
	if strings.HasPrefix(r.URL.Path, apiPath) {
		a.apiHandler.ServeHTTP(w, r)
		return renderer.Page{}, nil
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return renderer.NewPage("public", http.StatusOK, map[string]any{
		"Groups":    a.v2App.Groups(),
		"Scenes":    a.toScenes(),
		"Schedules": a.toSchedules(),
		"Sensors":   a.v2App.Sensors(),
	}), nil
}

func (a *App) toScenes() map[string]Scene {
	output := make(map[string]Scene, len(a.scenes))

	for key, item := range a.scenes {
		output[key] = item
	}

	return output
}

func (a *App) toSchedules() []Schedule {
	output := make([]Schedule, len(a.schedules))

	i := 0
	for _, item := range a.schedules {
		output[i] = item
		i++
	}

	sort.Sort(ByScheduleID(output))

	return output
}
