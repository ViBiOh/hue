package hue

import (
	"fmt"
	"strings"

	v2 "github.com/ViBiOh/hue/pkg/v2"
)

func getGroup(groups []v2.Group, name string) (v2.Group, error) {
	for _, group := range groups {
		if strings.EqualFold(group.Name, name) {
			return group, nil
		}
	}

	return v2.Group{}, fmt.Errorf("group `%s` not found", name)
}

func getTap(taps []v2.Tap, name string) (v2.Tap, error) {
	for _, tap := range taps {
		if strings.EqualFold(tap.Name, name) {
			return tap, nil
		}
	}

	return v2.Tap{}, fmt.Errorf("tap `%s` not found", name)
}
