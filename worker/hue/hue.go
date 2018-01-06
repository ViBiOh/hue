package hue

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/hue"
)

// App stores informations and secret of API
type App struct {
	bridgeURL string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	return &App{
		bridgeURL: getURL(*config[`bridgeIP`], *config[`username`]),
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`bridgeIP`: flag.String(tools.ToCamel(prefix+`bridgeIP`), ``, `[hue] IP of Bridge`),
		`username`: flag.String(tools.ToCamel(prefix+`username`), ``, `[hue] Username for Bridge`),
	}
}

func getURL(bridgeIP, username string) string {
	return `http://` + bridgeIP + `/api/` + username
}

func (a *App) getLight(lightID string) (*hue.Light, error) {
	content, err := httputils.GetBody(a.bridgeURL+`/lights/`+lightID, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting light from bridge: %v`, err)
	}

	var light hue.Light
	if err := json.Unmarshal(content, &light); err != nil {
		return nil, fmt.Errorf(`Error while parsing light data from bridge: %v`, err)
	}

	return &light, nil
}

func (a *App) getGroups() (map[string]*hue.Group, error) {
	content, err := httputils.GetBody(a.bridgeURL+`/groups`, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting groups from bridge: %v`, err)
	}

	var groups map[string]*hue.Group
	if err := json.Unmarshal(content, &groups); err != nil {
		return nil, fmt.Errorf(`Error while parsing groups from bridge: %v`, err)
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

// GetGroupsJSON get lists of groups in JSON
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
func (a *App) UpdateGroupState(groupID, state string) error {
	content, err := httputils.MethodBody(a.bridgeURL+`/groups/`+groupID+`/action`, []byte(state), nil, http.MethodPut)

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
