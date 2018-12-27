package hue

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/hue"
	"github.com/ViBiOh/iot/pkg/provider"
)

// Config of package
type Config struct {
	bridgeIP *string
	username *string
	config   *string
	clean    *bool
}

// App of package
type App struct {
	bridgeURL      string
	bridgeUsername string
	config         *hueConfig
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		bridgeIP: fs.String(tools.ToCamel(fmt.Sprintf(`%sBridgeIP`, prefix)), ``, `[hue] IP of Bridge`),
		username: fs.String(tools.ToCamel(fmt.Sprintf(`%sUsername`, prefix)), ``, `[hue] Username for Bridge`),
		config:   fs.String(tools.ToCamel(fmt.Sprintf(`%sConfig`, prefix)), ``, `[hue] Configuration filename`),
		clean:    fs.Bool(tools.ToCamel(fmt.Sprintf(`%sClean`, prefix)), false, `[hue] Clean Hue`),
	}
}

// New creates new App from Config
func New(config Config) (*App, error) {
	username := *config.username

	app := &App{
		bridgeUsername: username,
		bridgeURL:      fmt.Sprintf(`http://%s/api/%s`, *config.bridgeIP, username),
	}

	ctx := context.Background()

	if *config.clean {
		logger.Info(`Cleaning hue`)

		if err := app.cleanSchedules(ctx); err != nil {
			return nil, err
		}

		if err := app.cleanRules(ctx); err != nil {
			return nil, err
		}

		if err := app.cleanScenes(ctx); err != nil {
			return nil, err
		}
	}

	if *config.config != `` {
		rawConfig, err := ioutil.ReadFile(*config.config)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if err := json.Unmarshal(rawConfig, &app.config); err != nil {
			return nil, errors.WithStack(err)
		}

		app.configureSchedules(ctx, app.config.Schedules)
		app.configureTap(ctx, app.config.Taps)
		app.configureMotionSensor(ctx, app.config.Sensors)
	}

	return app, nil
}

func (a *App) handleStates(ctx context.Context, p *provider.WorkerMessage) error {
	if parts := strings.Split(p.Payload, `|`); len(parts) == 2 {
		state, ok := hue.States[parts[1]]
		if !ok {
			return errors.New(`unknown state %s`, parts[1])
		}

		if err := a.updateGroupState(ctx, parts[0], state); err != nil {
			return err
		}
	} else {
		return errors.New(`invalid state request: %s`, p.Payload)
	}

	return nil
}

func (a *App) handleSchedules(ctx context.Context, p *provider.WorkerMessage) error {
	if strings.HasSuffix(p.Action, hue.CreateAction) {
		var config hue.ScheduleConfig

		if err := json.Unmarshal([]byte(p.Payload), &config); err != nil {
			return errors.WithStack(err)
		}

		if err := a.createScheduleFromConfig(ctx, &config, nil); err != nil {
			return err
		}

		return nil
	}

	if strings.HasSuffix(p.Action, hue.UpdateAction) {
		var config hue.Schedule

		if err := json.Unmarshal([]byte(p.Payload), &config); err != nil {
			return errors.WithStack(err)
		}

		if config.ID == `` {
			return errors.New(`ID is missing`)
		}

		return a.updateSchedule(ctx, &config)
	}

	if strings.HasSuffix(p.Action, hue.DeleteAction) {
		id := p.Payload

		schedule, err := a.getSchedule(ctx, id)
		if err != nil {
			return err
		}

		if err := a.deleteSchedule(ctx, id); err != nil {
			return err
		}

		if sceneID, ok := schedule.Command.Body[`scene`]; ok {
			if err := a.deleteScene(ctx, sceneID.(string)); err != nil {
				return err
			}
		}

		return nil
	}

	return errors.New(`unknown schedule command`)
}

func (a *App) workerListGroups(ctx context.Context, initial *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	groups, err := a.listGroups(ctx)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(groups)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return provider.NewWorkerMessage(initial, hue.Source, hue.WorkerGroupsAction, fmt.Sprintf(`%s`, payload)), nil
}

func (a *App) workerListScenes(ctx context.Context, initial *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	scenes, err := a.listScenes(ctx)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(scenes)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return provider.NewWorkerMessage(initial, hue.Source, hue.WorkerScenesAction, fmt.Sprintf(`%s`, payload)), nil
}

func (a *App) workerListSchedules(ctx context.Context, initial *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	schedules, err := a.listSchedules(ctx)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(schedules)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return provider.NewWorkerMessage(initial, hue.Source, hue.WorkerSchedulesAction, fmt.Sprintf(`%s`, payload)), nil
}

func (a *App) workerListSensors(ctx context.Context, initial *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	sensors, err := a.listSensors(ctx)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(sensors)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return provider.NewWorkerMessage(initial, hue.Source, hue.WorkerSensorsAction, fmt.Sprintf(`%s`, payload)), nil
}

// Handle handle worker requests for Hue
func (a *App) Handle(ctx context.Context, p *provider.WorkerMessage) (*provider.WorkerMessage, error) {
	if strings.HasPrefix(p.Action, hue.WorkerGroupsAction) {
		return a.workerListGroups(ctx, p)
	}

	if strings.HasPrefix(p.Action, hue.WorkerScenesAction) {
		return a.workerListScenes(ctx, p)
	}

	if strings.HasPrefix(p.Action, hue.WorkerSensorsAction) {
		return a.workerListSensors(ctx, p)
	}

	if strings.HasPrefix(p.Action, hue.WorkerSchedulesAction) {
		if err := a.handleSchedules(ctx, p); err != nil {
			return nil, err
		}

		return a.workerListSchedules(ctx, p)
	}

	if strings.HasPrefix(p.Action, hue.WorkerStateAction) {
		if err := a.handleStates(ctx, p); err != nil {
			return nil, err
		}

		return a.workerListGroups(ctx, p)
	}

	return nil, errors.New(`unknown request: %s`, p)
}

// GetSource returns source name
func (a *App) GetSource() string {
	return hue.Source
}

// Ping send to worker updated data
func (a *App) Ping(ctx context.Context) ([]*provider.WorkerMessage, error) {
	groups, err := a.workerListGroups(ctx, nil)
	if err != nil {
		return nil, err
	}

	scenes, err := a.workerListScenes(ctx, nil)
	if err != nil {
		return nil, err
	}

	schedules, err := a.workerListSchedules(ctx, nil)
	if err != nil {
		return nil, err
	}

	sensors, err := a.workerListSensors(ctx, nil)
	if err != nil {
		return nil, err
	}

	return []*provider.WorkerMessage{groups, scenes, schedules, sensors}, nil
}
