package v2

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// Group description
type Group struct {
	GroupedLights map[string]GroupedLight
	ID            string `json:"id"`
	Name          string `json:"name"`
	Lights        []*Light
}

// GroupedLight description
type GroupedLight struct {
	ID    string `json:"id"`
	Alert struct {
		ActionValues []string `json:"action_values"`
	} `json:"alert"`
	Dimming Dimming `json:"dimming"`
	On      On      `json:"on"`
}

// Room description
type Room struct {
	ID       string `json:"id"`
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Services []deviceReference `json:"services"`
	Children []deviceReference `json:"children"`
}

func (a *App) buildGroup(ctx context.Context) (output map[string]Group, err error) {
	output = make(map[string]Group)

	err = a.buildDeviceGroup(ctx, "room", output)
	if err != nil {
		return
	}

	err = a.buildDeviceGroup(ctx, "zone", output)
	if err != nil {
		return
	}

	err = a.buildDeviceGroup(ctx, "bridge_home", output)
	if err != nil {
		return
	}

	return
}

func (a *App) buildDeviceGroup(ctx context.Context, name string, output map[string]Group) error {
	groupDevices, err := list[Room](ctx, a.req, name)
	if err != nil {
		return fmt.Errorf("unable to list rooms: %s", err)
	}

	for _, item := range groupDevices {
		groupedLights, err := a.buildServices(ctx, name, item.Services)
		if err != nil {
			return fmt.Errorf("unable to build services for %s `%s`: %s", name, item.ID, err)
		}

		lights, err := a.buildChildren(ctx, name, item.Children)
		if err != nil {
			return fmt.Errorf("unable to build children for %s `%s`: %s", name, item.ID, err)
		}

		output[item.ID] = Group{
			ID:            item.ID,
			Name:          item.Metadata.Name,
			GroupedLights: groupedLights,
			Lights:        lights,
		}
	}

	return nil
}

func (a *App) buildServices(ctx context.Context, name string, services []deviceReference) (map[string]GroupedLight, error) {
	output := make(map[string]GroupedLight)

	for _, service := range services {
		switch service.Rtype {
		case "grouped_light":
			groupedLight, err := get[GroupedLight](ctx, a.req, service.Rtype, service.Rid)
			if err != nil {
				return nil, fmt.Errorf("unable to get grouped light `%s`: %s", service.Rid, err)
			}
			output[groupedLight.ID] = groupedLight

		default:
			logger.Warn("unhandled service type for %s: %s", name, service.Rtype)
		}
	}

	return output, nil
}

func (a *App) buildChildren(ctx context.Context, name string, children []deviceReference) ([]*Light, error) {
	var output []*Light

	for _, service := range children {
		switch service.Rtype {
		case "light":
			if light, ok := a.lights[service.Rid]; ok {
				output = append(output, light)
			}
		case "device":
			device, err := get[Device](ctx, a.req, service.Rtype, service.Rid)
			if err != nil {
				return nil, fmt.Errorf("unable to get device `%s`: %s", service.Rid, err)
			}

			lights, err := a.buildChildren(ctx, "device", device.Services)
			if err != nil {
				return nil, fmt.Errorf("unable to get children of device `%s`: %s", service.Rid, err)
			}

			output = append(output, lights...)
		}
	}

	return output, nil
}

func (a *App) getGroupOfGroupedLight(groupedLightID string) (Group, bool) {
	for _, group := range a.groups {
		if _, ok := group.GroupedLights[groupedLightID]; ok {
			return group, true
		}
	}

	return Group{}, false
}
