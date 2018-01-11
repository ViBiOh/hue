package hue

import (
	"fmt"
	"strings"

	"github.com/ViBiOh/iot/hue"
)

func (a *App) listGroups() (map[string]*hue.Group, error) {
	var groups map[string]*hue.Group
	err := get(fmt.Sprintf(`%s/groups`, a.bridgeURL), &groups)
	if err != nil {
		return nil, err
	}

	for _, value := range groups {
		value.Tap = false

		for _, lightID := range value.Lights {
			light, err := a.getLight(lightID)
			if err != nil {
				return nil, fmt.Errorf(`Error while getting light data of group: %v`, err)
			}

			if strings.HasPrefix(light.Type, `On/Off`) {
				value.Tap = true
			}
		}
	}

	return groups, nil
}

func (a *App) updateGroupState(groupID string, state interface{}) error {
	return update(fmt.Sprintf(`%s/groups/%s/action`, a.bridgeURL, groupID), state)
}
