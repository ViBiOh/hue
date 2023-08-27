package v2

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"time"
)

// Group description
type Group struct {
	GroupedLights map[string]GroupedLight
	ID            string
	Name          string
	Lights        []*Light
	Bridge        bool
	Plug          bool
}

// AnyOn checks if any lights in the group is on
func (g Group) AnyOn() bool {
	for _, light := range g.Lights {
		if light.On.On {
			return true
		}
	}

	return false
}

func isPlug(lights []*Light) bool {
	var count int

	for _, light := range lights {
		if light.Metadata.Archetype == "plug" {
			count++
		}
	}

	return count > 0 && count == len(lights)
}

// GroupByTypeAndName sort Group by Type then, Name
type GroupByTypeAndName []Group

func (a GroupByTypeAndName) Len() int      { return len(a) }
func (a GroupByTypeAndName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a GroupByTypeAndName) Less(i, j int) bool {
	if a[i].Plug == a[j].Plug {
		return a[i].Name < a[j].Name
	}

	if a[i].Plug && !a[j].Plug {
		return false
	}
	return true
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

// Groups list available groups
func (s *Service) Groups() []Group {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	output := make([]Group, len(s.groups))

	i := 0
	for _, item := range s.groups {
		output[i] = item
		i++
	}

	sort.Sort(GroupByTypeAndName(output))

	return output
}

// UpdateGroup status
func (s *Service) UpdateGroup(ctx context.Context, id string, on bool, brightness float64, transitionTime time.Duration) (Group, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// var color Color
	// color.XY.X = 0.313
	// color.XY.Y = .337

	var colorTemperature ColorTemperature
	colorTemperature.Mirek = 239

	payload := map[string]interface{}{
		"on": On{
			On: on,
		},
		"dimming": Dimming{
			Brightness: brightness,
		},
		"color_temperature": colorTemperature,
		"dynamics": map[string]interface{}{
			"duration": transitionTime.Milliseconds(),
		},
	}

	group, ok := s.groups[id]
	if !ok {
		return group, fmt.Errorf("unknown group with id `%s`", id)
	}

	for _, groupedLight := range group.GroupedLights {
		if _, err := s.req.Method(http.MethodPut).Path("/clip/v2/resource/grouped_light/"+groupedLight.ID).JSON(ctx, payload); err != nil {
			return group, fmt.Errorf("update grouped light `%s`: %w", groupedLight.ID, err)
		}
	}

	return group, nil
}

func (s *Service) buildGroup(ctx context.Context) (output map[string]Group, err error) {
	output = make(map[string]Group)

	err = s.buildDeviceGroup(ctx, "room", output)
	if err != nil {
		return
	}

	err = s.buildDeviceGroup(ctx, "zone", output)
	if err != nil {
		return
	}

	err = s.buildDeviceGroup(ctx, "bridge_home", output)
	if err != nil {
		return
	}

	return
}

func (s *Service) buildDeviceGroup(ctx context.Context, name string, output map[string]Group) error {
	groupDevices, err := list[Room](ctx, s.req, name)
	if err != nil {
		return fmt.Errorf("list rooms: %w", err)
	}

	isBridge := name == "bridge_home"

	for _, item := range groupDevices {
		groupedLights, err := s.buildServices(ctx, name, item.Services)
		if err != nil {
			return fmt.Errorf("build services for %s `%s`: %w", name, item.ID, err)
		}

		lights, err := s.buildChildren(ctx, item.Children)
		if err != nil {
			return fmt.Errorf("build children for %s `%s`: %w", name, item.ID, err)
		}

		groupName := item.Metadata.Name
		if isBridge {
			groupName = "Bridge"
		}

		output[item.ID] = Group{
			ID:            item.ID,
			Name:          groupName,
			GroupedLights: groupedLights,
			Lights:        lights,
			Plug:          isPlug(lights),
			Bridge:        isBridge,
		}
	}

	return nil
}

func (s *Service) buildServices(ctx context.Context, name string, services []deviceReference) (map[string]GroupedLight, error) {
	output := make(map[string]GroupedLight)

	for _, service := range services {
		switch service.Rtype {
		case "grouped_light":
			groupedLight, err := get[GroupedLight](ctx, s.req, service.Rtype, service.Rid)
			if err != nil {
				return nil, fmt.Errorf("get grouped light `%s`: %w", service.Rid, err)
			}
			output[groupedLight.ID] = groupedLight

		default:
			slog.Warn("unhandled service type", "anem", name, "type", service.Rtype)
		}
	}

	return output, nil
}

func (s *Service) buildChildren(ctx context.Context, children []deviceReference) ([]*Light, error) {
	var output []*Light

	for _, service := range children {
		switch service.Rtype {
		case "light":
			if light, ok := s.lights[service.Rid]; ok {
				output = append(output, light)
			}
		case "device":
			device, err := get[Device](ctx, s.req, service.Rtype, service.Rid)
			if err != nil {
				return nil, fmt.Errorf("get device `%s`: %w", service.Rid, err)
			}

			lights, err := s.buildChildren(ctx, device.Services)
			if err != nil {
				return nil, fmt.Errorf("get children of device `%s`: %w", service.Rid, err)
			}

			output = append(output, lights...)
		}
	}

	return output, nil
}

func (s *Service) getGroupOfGroupedLight(groupedLightID string) (Group, bool) {
	for _, group := range s.groups {
		if _, ok := group.GroupedLights[groupedLightID]; ok {
			return group, true
		}
	}

	return Group{}, false
}
