package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/ViBiOh/httputils"
)

const off = `{"on":false,"transitiontime":30}`
const dimmed = `{"on":true,"transitiontime":30,"sat":0,"bri":0}`
const bright = `{"on":true,"transitiontime":30,"sat":0,"bri":254}`

type light struct {
	Name  string
	State struct {
		On bool
	}
}

func getURL(bridgeIP, username string) string {
	return `http://` + bridgeIP + `/api/` + username + `/lights`
}

func listLights(bridgeIP, username string) ([]light, error) {
	content, err := httputils.GetBody(getURL(bridgeIP, username), nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting data from bridge: %v`, err)
	}

	var rawLights map[string]light
	if err := json.Unmarshal(content, &rawLights); err != nil {
		return nil, fmt.Errorf(`Error while parsing data from bridge: %v`, err)
	}

	lights := make([]light, len(rawLights))
	for key, value := range rawLights {
		i, _ := strconv.Atoi(key)
		lights[i-1] = value
	}

	return lights, nil
}

func updateState(bridgeIP, username, light, state string) error {
	content, err := httputils.MethodBody(getURL(bridgeIP, username)+`/`+light+`/state`, []byte(state), nil, http.MethodPut)

	if err != nil {
		return fmt.Errorf(`Error while sending data to bridge: %v`, err)
	}

	if bytes.Contains(content, []byte(`error`)) {
		return fmt.Errorf(`Error while updating state: %s`, content)
	}

	return nil
}

func updateAllState(bridgeIP, username, state string) error {
	lights, err := listLights(bridgeIP, username)
	if err != nil {
		return fmt.Errorf(`Error while listing lights: %v`, err)
	}

	for index, light := range lights {
		if err := updateState(bridgeIP, username, strconv.Itoa(index+1), state); err != nil {
			return fmt.Errorf(`Error while updating %s: %v`, light.Name, err)
		}
	}

	return nil
}

func main() {
	bridgeIP := flag.String(`bridgeIP`, ``, `IP of Hue Bridge`)
	username := flag.String(`username`, ``, `Username for Hue Bridge`)
	flag.Parse()

	log.Print(updateAllState(*bridgeIP, *username, bright))
}
