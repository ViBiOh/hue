package hue

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/hue"
)

// App stores informations and secret of API
type App struct {
	bridgeURL string
	tap       *tapConfig
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}) (*App, error) {
	app := &App{
		bridgeURL: getURL(*config[`bridgeIP`].(*string), *config[`username`].(*string)),
	}

	if *config[`clean`].(*bool) {
		app.cleanSchedules()
		app.cleanScenes()
		app.cleanRules()
	}

	if *config[`tapConfig`].(*string) != `` {
		rawTapConfig, err := ioutil.ReadFile(*config[`tapConfig`].(*string))
		if err != nil {
			return nil, fmt.Errorf(`Error while reading tap config filename: %v`, err)
		}

		if err := json.Unmarshal(rawTapConfig, &app.tap); err != nil {
			return nil, fmt.Errorf(`Error while unmarshalling tap config: %v`, err)
		}

		app.configureTap()
	}

	return app, nil
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`bridgeIP`:  flag.String(tools.ToCamel(prefix+`BridgeIP`), ``, `[hue] IP of Bridge`),
		`username`:  flag.String(tools.ToCamel(prefix+`Username`), ``, `[hue] Username for Bridge`),
		`tapConfig`: flag.String(tools.ToCamel(prefix+`TapConfig`), ``, `[hue] Tap configuration filename`),
		`clean`:     flag.Bool(tools.ToCamel(prefix+`Clean`), false, `[hue] Clean Hue`),
	}
}

func getURL(bridgeIP, username string) string {
	return `http://` + bridgeIP + `/api/` + username
}

func (a *App) getLight(lightID string) (*hue.Light, error) {
	content, err := httputils.GetRequest(a.bridgeURL+`/lights/`+lightID, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting light: %v`, err)
	}

	var light hue.Light
	if err := json.Unmarshal(content, &light); err != nil {
		return nil, fmt.Errorf(`Error while parsing light data: %v`, err)
	}

	return &light, nil
}

func (a *App) getGroups() (map[string]*hue.Group, error) {
	content, err := httputils.GetRequest(a.bridgeURL+`/groups`, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting groups: %v`, err)
	}

	var groups map[string]*hue.Group
	if err := json.Unmarshal(content, &groups); err != nil {
		return nil, fmt.Errorf(`Error while parsing groups: %v`, err)
	}

	for _, value := range groups {
		value.OnOff = true
		value.On = false

		for _, lightID := range value.Lights {
			light, err := a.getLight(lightID)
			if err != nil {
				return nil, fmt.Errorf(`Error while getting light data of group: %v`, err)
			}

			if !strings.HasPrefix(light.Type, `On/Off`) {
				value.OnOff = false
			}
			if light.State.On {
				value.On = true
			}
		}
	}

	return groups, nil
}

// GetGroupsPayload get lists of groups in websocket format
func (a *App) GetGroupsPayload() ([]byte, error) {
	groups, err := a.getGroups()
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

// UpdateGroupState update state of group
func (a *App) UpdateGroupState(groupID string, state map[string]interface{}) error {
	content, err := httputils.RequestJSON(fmt.Sprintf(`%s/groups/%s/action`, a.bridgeURL, groupID), state, nil, http.MethodPut)

	if err != nil {
		return fmt.Errorf(`Error while sending data to bridge: %v`, err)
	}

	if bytes.Contains(content, []byte(`error`)) {
		return fmt.Errorf(`Error while updating state: %s`, content)
	}

	return nil
}

// Handle handle worker requests for Hue
func (a *App) Handle(p []byte) ([]byte, error) {
	if bytes.HasPrefix(p, hue.GroupsPrefix) {
		return a.GetGroupsPayload()
	} else if bytes.HasPrefix(p, hue.StatePrefix) {
		if parts := bytes.Split(bytes.TrimPrefix(p, hue.StatePrefix), []byte(`|`)); len(parts) == 2 {
			state, ok := hue.States[string(parts[1])]
			if !ok {
				return nil, fmt.Errorf(`Unknown state %s`, parts[1])
			}

			if err := a.UpdateGroupState(string(parts[0]), state); err != nil {
				return nil, err
			}
			return a.GetGroupsPayload()
		}
	}

	return nil, fmt.Errorf(`Unknown request: %s`, p)
}
