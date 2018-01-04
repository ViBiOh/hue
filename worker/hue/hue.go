package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/iot/hue"
)

// GetURL forge bridge URL for bridge API
func GetURL(bridgeIP, username string) string {
	return `http://` + bridgeIP + `/api/` + username
}

func getLight(bridgeURL, lightID string) (*hue.Light, error) {
	content, err := httputils.GetBody(bridgeURL+`/lights/`+lightID, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting light from bridge: %v`, err)
	}

	var light hue.Light
	if err := json.Unmarshal(content, &light); err != nil {
		return nil, fmt.Errorf(`Error while parsing light data from bridge: %v`, err)
	}

	return &light, nil
}

func getGroups(bridgeURL string) (map[string]*hue.Group, error) {
	content, err := httputils.GetBody(bridgeURL+`/groups`, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting groups from bridge: %v`, err)
	}

	var groups map[string]*hue.Group
	if err := json.Unmarshal(content, &groups); err != nil {
		return nil, fmt.Errorf(`Error while parsing groups from bridge: %v`, err)
	}

	for _, value := range groups {
		value.OnOff = true

		for _, lightID := range value.Lights {
			light, err := getLight(bridgeURL, lightID)
			if err != nil {
				return nil, fmt.Errorf(`Error while getting light data of group: %v`, err)
			}

			if !strings.HasPrefix(light.Type, `On/Off`) {
				value.OnOff = false
				break
			}
		}
	}

	return groups, nil
}

// GetGroupsJSON get lists of groups in JSON
func GetGroupsJSON(bridgeURL string) ([]byte, error) {
	groups, err := getGroups(bridgeURL)
	if err != nil {
		err = fmt.Errorf(`Error while listing groups: %v`, err)
		return nil, err
	}

	groupsJSON, err := json.Marshal(groups)
	if err != nil {
		err = fmt.Errorf(`Error while marshalling groups: %v`, err)
		return nil, err
	}

	return groupsJSON, nil
}

// UpdateGroupState update state of group
func UpdateGroupState(bridgeURL, groupID, state string) error {
	content, err := httputils.MethodBody(bridgeURL+`/groups/`+groupID+`/action`, []byte(state), nil, http.MethodPut)

	if err != nil {
		return fmt.Errorf(`Error while sending data to bridge: %v`, err)
	}

	if bytes.Contains(content, []byte(`error`)) {
		return fmt.Errorf(`Error while updating state: %s`, content)
	}

	return nil
}
