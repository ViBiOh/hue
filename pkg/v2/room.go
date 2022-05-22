package v2

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// Group description
type Group struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	GroupedLights map[string]GroupedLight
}

// Dimming description
type Dimming struct {
	Brightness float64 `json:"brightness"`
}

// On description
type On struct {
	On bool `json:"on"`
}

// GroupedLight description
type GroupedLight struct {
	Alert struct {
		ActionValues []string `json:"action_values"`
	} `json:"alert"`
	Dimming Dimming `json:"dimming"`
	ID      string  `json:"id"`
	On      On      `json:"on"`
}

// Service description
type Service struct {
	Rid   string `json:"rid"`
	Rtype string `json:"rtype"`
}

// Room description
type Room struct {
	ID       string `json:"id"`
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Services []Service `json:"services"`
}

// BridgeHome description
type BridgeHome struct {
	ID       string    `json:"id"`
	Services []Service `json:"services"`
}

func (a *App) buildGroup(ctx context.Context) (map[string]Group, error) {
	rooms, err := list[Room](ctx, a.req, "room")
	if err != nil {
		return nil, fmt.Errorf("unable to list rooms: %s", err)
	}

	output := make(map[string]Group, len(rooms))
	for _, room := range rooms {
		groupedLights, err := a.buildServices(ctx, room.Services)
		if err != nil {
			return nil, fmt.Errorf("unable to build services for room `%s`: %s", room.ID, err)
		}

		output[room.ID] = Group{
			ID:            room.ID,
			Name:          room.Metadata.Name,
			GroupedLights: groupedLights,
		}
	}

	bridgeHomes, err := list[BridgeHome](ctx, a.req, "bridge_home")
	if err != nil {
		return nil, fmt.Errorf("unable to list bridge homes: %s", err)
	}

	for _, bridgeHome := range bridgeHomes {
		groupedLights, err := a.buildServices(ctx, bridgeHome.Services)
		if err != nil {
			return nil, fmt.Errorf("unable to build services for bridge `%s`: %s", bridgeHome.ID, err)
		}

		output[bridgeHome.ID] = Group{
			ID:            bridgeHome.ID,
			Name:          "Bridge",
			GroupedLights: groupedLights,
		}
	}

	return output, nil
}

func (a *App) buildServices(ctx context.Context, services []Service) (map[string]GroupedLight, error) {
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
			logger.Warn("unknown room's service type: %s", service.Rtype)
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
