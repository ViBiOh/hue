package hue

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/hue"
	"github.com/ViBiOh/iot/provider"
)

var debug = false

// App stores informations and secret of API
type App struct {
	bridgeURL      string
	bridgeUsername string
	config         *hueConfig
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}) (*App, error) {
	if *config[`debug`].(*bool) {
		debug = true
	}

	username := *config[`username`].(*string)

	app := &App{
		bridgeUsername: username,
		bridgeURL:      fmt.Sprintf(`http://%s/api/%s`, *config[`bridgeIP`].(*string), username),
	}

	if *config[`clean`].(*bool) {
		if err := app.cleanSchedules(); err != nil {
			return nil, fmt.Errorf(`Error while cleaning schedules: %v`, err)
		}

		if err := app.cleanScenes(); err != nil {
			return nil, fmt.Errorf(`Error while cleaning scenes: %v`, err)
		}

		if err := app.cleanRules(); err != nil {
			return nil, fmt.Errorf(`Error while cleaning rules: %v`, err)
		}
	}

	if *config[`config`].(*string) != `` {
		rawConfig, err := ioutil.ReadFile(*config[`config`].(*string))
		if err != nil {
			return nil, fmt.Errorf(`Error while reading config filename: %v`, err)
		}

		if err := json.Unmarshal(rawConfig, &app.config); err != nil {
			return nil, fmt.Errorf(`Error while unmarshalling config %s: %v`, rawConfig, err)
		}

		app.configureSchedules(app.config.Schedules)
		app.configureTap(app.config.Taps)
	}

	return app, nil
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`bridgeIP`: flag.String(tools.ToCamel(prefix+`BridgeIP`), ``, `[hue] IP of Bridge`),
		`username`: flag.String(tools.ToCamel(prefix+`Username`), ``, `[hue] Username for Bridge`),
		`config`:   flag.String(tools.ToCamel(prefix+`Config`), ``, `[hue] Configuration filename`),
		`clean`:    flag.Bool(tools.ToCamel(prefix+`Clean`), false, `[hue] Clean Hue`),
		`debug`:    flag.Bool(tools.ToCamel(prefix+`Debug`), false, `Enable debug logging`),
	}
}

func (a *App) formatWorkerMessage(initial *provider.WorkerMessage, messageType string, payload interface{}) *provider.WorkerMessage {
	return &provider.WorkerMessage{
		ID:      initial.ID,
		Source:  hue.HueSource,
		Type:    messageType,
		Payload: payload,
	}
}

// Handle handle worker requests for Hue
func (a *App) Handle(p *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	if strings.HasPrefix(p.Type, hue.GroupsPrefix) {
		output, err := a.listGroups()
		if err != nil {
			return nil, err
		}
		return a.formatWorkerMessage(p, hue.GroupsPrefix, output), nil
	}

	if strings.HasPrefix(p.Type, hue.ScenesPrefix) {
		output, err := a.listScenes()
		if err != nil {
			return nil, err
		}

		return a.formatWorkerMessage(p, hue.ScenesPrefix, output), nil
	}

	if strings.HasPrefix(p.Type, hue.SchedulesPrefix) {
		if strings.HasSuffix(p.Type, hue.CreatePrefix) {
			var config hue.ScheduleConfig

			if convert, err := json.Marshal(p.Payload); err != nil {
				return nil, fmt.Errorf(`Error while converting schedules payload: %v`, err)
			} else if err := json.Unmarshal(convert, &config); err != nil {
				return nil, fmt.Errorf(`Error while unmarshalling schedule create config: %v`, err)
			}

			if err := a.createScheduleFromConfig(&config, nil); err != nil {
				return nil, fmt.Errorf(`Error while creating schedule from config: %v`, err)
			}
		} else if strings.HasSuffix(p.Type, hue.UpdatePrefix) {
			var config hue.Schedule

			if convert, err := json.Marshal(p.Payload); err != nil {
				return nil, fmt.Errorf(`Error while converting schedules payload: %v`, err)
			} else if err := json.Unmarshal(convert, &config); err != nil {
				return nil, fmt.Errorf(`Error while unmarshalling schedule update: %v`, err)
			}

			if config.ID == `` {
				return nil, errors.New(`Error while updating schedule config: ID is missing`)
			}

			if err := a.updateSchedule(&config); err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(p.Type, hue.DeletePrefix) {
			id := string(p.Payload.([]byte))

			schedule, err := a.getSchedule(id)
			if err != nil {
				return nil, fmt.Errorf(`Error while getting schedule: %v`, err)
			}

			if err := a.deleteSchedule(id); err != nil {
				return nil, fmt.Errorf(`Error while deleting schedule: %v`, err)
			}

			if sceneID, ok := schedule.Command.Body[`scene`]; ok {
				if err := a.deleteScene(sceneID.(string)); err != nil {
					return nil, fmt.Errorf(`Error while deleting scene: %v`, err)
				}
			}
		}

		output, err := a.listSchedules()
		if err != nil {
			return nil, err
		}

		return a.formatWorkerMessage(p, hue.SchedulesPrefix, output), nil
	}

	if strings.HasPrefix(p.Type, hue.StatePrefix) {
		if parts := bytes.Split(p.Payload.([]byte), []byte(`|`)); len(parts) == 2 {
			state, ok := hue.States[string(parts[1])]
			if !ok {
				return nil, fmt.Errorf(`Unknown state %s`, parts[1])
			}

			if err := a.updateGroupState(string(parts[0]), state); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf(`Invalid state request: %s`, p.Payload)
		}

		output, err := a.listGroups()
		if err != nil {
			return nil, err
		}

		return a.formatWorkerMessage(p, hue.GroupsPrefix, output), nil
	}

	return nil, fmt.Errorf(`Unknown request: %s`, p)
}
