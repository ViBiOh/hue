package hue

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/hue"
	"github.com/ViBiOh/iot/pkg/provider"
)

var debug = false

// App stores informations and secret of API
type App struct {
	bridgeURL      string
	bridgeUsername string
	config         *hueConfig
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}, debugApp bool) (*App, error) {
	debug = debugApp

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
		`bridgeIP`: flag.String(tools.ToCamel(fmt.Sprintf(`%sBridgeIP`, prefix)), ``, `[hue] IP of Bridge`),
		`username`: flag.String(tools.ToCamel(fmt.Sprintf(`%sUsername`, prefix)), ``, `[hue] Username for Bridge`),
		`config`:   flag.String(tools.ToCamel(fmt.Sprintf(`%sConfig`, prefix)), ``, `[hue] Configuration filename`),
		`clean`:    flag.Bool(tools.ToCamel(fmt.Sprintf(`%sClean`, prefix)), false, `[hue] Clean Hue`),
	}
}

func (a *App) formatWorkerMessage(initial *provider.WorkerMessage, messageType string, payload interface{}) *provider.WorkerMessage {
	id := ``
	if initial != nil {
		id = initial.ID
	}

	return &provider.WorkerMessage{
		ID:      id,
		Source:  hue.HueSource,
		Type:    messageType,
		Payload: payload,
	}
}

func (a *App) handleStates(p *provider.WorkerMessage) error {
	if parts := strings.Split(p.Payload.(string), `|`); len(parts) == 2 {
		state, ok := hue.States[parts[1]]
		if !ok {
			return fmt.Errorf(`Unknown state %s`, parts[1])
		}

		if err := a.updateGroupState(parts[0], state); err != nil {
			return err
		}
	} else {
		return fmt.Errorf(`Invalid state request: %s`, p.Payload)
	}

	return nil
}

func (a *App) handleSchedules(p *provider.WorkerMessage) error {
	if strings.HasSuffix(p.Type, hue.CreateAction) {
		var config hue.ScheduleConfig

		if err := json.Unmarshal([]byte(p.Payload.(string)), &config); err != nil {
			return fmt.Errorf(`Error while unmarshalling schedule create config: %v`, err)
		}

		if err := a.createScheduleFromConfig(&config, nil); err != nil {
			return fmt.Errorf(`Error while creating schedule from config: %v`, err)
		}

		return nil
	}

	if strings.HasSuffix(p.Type, hue.UpdateAction) {
		var config hue.Schedule

		if err := json.Unmarshal([]byte(p.Payload.(string)), &config); err != nil {
			return fmt.Errorf(`Error while unmarshalling schedule update: %v`, err)
		}

		if config.ID == `` {
			return errors.New(`Error while updating schedule config: ID is missing`)
		}

		return a.updateSchedule(&config)
	}

	if strings.HasSuffix(p.Type, hue.DeleteAction) {
		id := p.Payload.(string)

		schedule, err := a.getSchedule(id)
		if err != nil {
			return fmt.Errorf(`Error while getting schedule: %v`, err)
		}

		if err := a.deleteSchedule(id); err != nil {
			return fmt.Errorf(`Error while deleting schedule: %v`, err)
		}

		if sceneID, ok := schedule.Command.Body[`scene`]; ok {
			if err := a.deleteScene(sceneID.(string)); err != nil {
				return fmt.Errorf(`Error while deleting scene: %v`, err)
			}
		}

		return nil
	}

	return errors.New(`Unknown schedule command`)
}

func (a *App) workerListGroups(initial *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	output, err := a.listGroups()
	if err != nil {
		return nil, err
	}
	return a.formatWorkerMessage(initial, hue.WorkerGroupsType, output), nil
}

func (a *App) workerListScenes(initial *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	output, err := a.listScenes()
	if err != nil {
		return nil, err
	}
	return a.formatWorkerMessage(initial, hue.WorkerScenesType, output), nil
}

func (a *App) workerListSchedules(initial *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	output, err := a.listSchedules()
	if err != nil {
		return nil, err
	}
	return a.formatWorkerMessage(initial, hue.WorkerSchedulesType, output), nil
}

// Handle handle worker requests for Hue
func (a *App) Handle(p *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	if strings.HasPrefix(p.Type, hue.WorkerGroupsType) {
		return a.workerListGroups(p)
	}

	if strings.HasPrefix(p.Type, hue.WorkerScenesType) {
		return a.workerListScenes(p)
	}

	if strings.HasPrefix(p.Type, hue.WorkerSchedulesType) {
		if err := a.handleSchedules(p); err != nil {
			return nil, err
		}

		return a.workerListSchedules(p)
	}

	if strings.HasPrefix(p.Type, hue.WorkerStateType) {
		if err := a.handleStates(p); err != nil {
			return nil, err
		}

		return a.workerListGroups(p)
	}

	return nil, fmt.Errorf(`Unknown request: %s`, p)
}

// Ping send to worker update informations
func (a *App) Ping() ([]*provider.WorkerMessage, error) {
	groups, err := a.workerListGroups(nil)
	if err != nil {
		return nil, err
	}

	scenes, err := a.workerListScenes(nil)
	if err != nil {
		return nil, err
	}

	schedules, err := a.workerListSchedules(nil)
	if err != nil {
		return nil, err
	}

	return []*provider.WorkerMessage{groups, scenes, schedules}, nil
}
