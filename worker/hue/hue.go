package hue

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/hue"
)

// App stores informations and secret of API
type App struct {
	bridgeURL      string
	bridgeUsername string
	config         *hueConfig
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}) (*App, error) {
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
			return nil, fmt.Errorf(`Error while unmarshalling config: %v`, err)
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
	}
}

// GetGroupsPayload get lists of groups in websocket format
func (a *App) GetGroupsPayload() ([]byte, error) {
	groups, err := a.listGroups()
	if err != nil {
		err = fmt.Errorf(`Error while listing groups: %v`, err)
		return nil, err
	}

	groupsJSON, err := json.Marshal(groups)
	if err != nil {
		err = fmt.Errorf(`Error while marshalling groups: %v`, err)
		return nil, err
	}

	return append(hue.GroupsPrefix, groupsJSON...), nil
}

// GetSchedulesPayload get lists of schedules in websocket format
func (a *App) GetSchedulesPayload() ([]byte, error) {
	schedules, err := a.listSchedules()
	if err != nil {
		err = fmt.Errorf(`Error while listing schedules: %v`, err)
		return nil, err
	}

	schedulesJSON, err := json.Marshal(schedules)
	if err != nil {
		err = fmt.Errorf(`Error while marshalling schedules: %v`, err)
		return nil, err
	}

	return append(hue.SchedulesPrefix, schedulesJSON...), nil
}

// Handle handle worker requests for Hue
func (a *App) Handle(p []byte) ([]byte, error) {
	if bytes.HasPrefix(p, hue.GroupsPrefix) {
		return a.GetGroupsPayload()
	}

	if bytes.HasPrefix(p, hue.SchedulesPrefix) {
		request := bytes.TrimPrefix(p, hue.SchedulesPrefix)

		if bytes.HasPrefix(request, hue.CreatePrefix) {
			var config *hue.ScheduleConfig
			if err := json.Unmarshal(bytes.TrimPrefix(request, hue.CreatePrefix), config); err != nil {
				return nil, fmt.Errorf(`Error while unmarshalling schedule create config: %v`, err)
			}

			if err := a.createScheduleFromConfig(config, nil); err != nil {
				return nil, fmt.Errorf(`Error while creating schedule from config: %v`, err)
			}
		} else if bytes.HasPrefix(request, hue.UpdatePrefix) {
			var config *hue.Schedule
			if err := json.Unmarshal(bytes.TrimPrefix(request, hue.UpdatePrefix), config); err != nil {
				return nil, fmt.Errorf(`Error while unmarshalling schedule create config: %v`, err)
			}

			if config.ID == `` {
				return nil, errors.New(`Error while updating schedule config: ID is missing`)
			}

			if err := a.updateSchedule(config); err != nil {
				return nil, err
			}
		}

		return a.GetSchedulesPayload()
	}

	if bytes.HasPrefix(p, hue.StatePrefix) {
		request := bytes.TrimPrefix(p, hue.StatePrefix)

		if parts := bytes.Split(request, []byte(`|`)); len(parts) == 2 {
			state, ok := hue.States[string(parts[1])]
			if !ok {
				return nil, fmt.Errorf(`Unknown state %s`, parts[1])
			}

			if err := a.updateGroupState(string(parts[0]), state); err != nil {
				return nil, err
			}
		}

		return a.GetGroupsPayload()
	}

	return nil, fmt.Errorf(`Unknown request: %s`, p)
}
