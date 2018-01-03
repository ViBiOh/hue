package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
	"github.com/gorilla/websocket"
)

const pingDelay = 60 * time.Second

var states = map[string]string{
	`off`:    `{"on":false,"transitiontime":30}`,
	`dimmed`: `{"on":true,"transitiontime":30,"sat":0,"bri":0}`,
	`bright`: `{"on":true,"transitiontime":30,"sat":0,"bri":254}`,
}

type light struct {
	Name  string
	State struct {
		On bool
	}
}

func getURL(bridgeIP, username string) string {
	return `http://` + bridgeIP + `/api/` + username + `/lights`
}

func listLights(bridgeURL string) ([]light, error) {
	content, err := httputils.GetBody(bridgeURL, nil)
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

func updateState(bridgeURL, light, state string) error {
	content, err := httputils.MethodBody(bridgeURL+`/`+light+`/state`, []byte(state), nil, http.MethodPut)

	if err != nil {
		return fmt.Errorf(`Error while sending data to bridge: %v`, err)
	}

	if bytes.Contains(content, []byte(`error`)) {
		return fmt.Errorf(`Error while updating state: %s`, content)
	}

	return nil
}

func updateAllState(bridgeURL, state string) error {
	lights, err := listLights(bridgeURL)
	if err != nil {
		return fmt.Errorf(`Error while listing lights: %v`, err)
	}

	for index, light := range lights {
		if err := updateState(bridgeURL, strconv.Itoa(index+1), state); err != nil {
			return fmt.Errorf(`Error while updating %s: %v`, light.Name, err)
		}
	}

	return nil
}

func connect(url string, bridgeURL string, secretKey string) {
	localIPS, err := tools.GetLocalIPS()
	if err != nil {
		log.Printf(`Error while retrieving local ips: %v`, err)
		return
	}

	headers := http.Header{}
	headers.Set(httputils.ForwardedForHeader, localIPS[0].String())

	ws, _, err := websocket.DefaultDialer.Dial(url, headers)
	if ws != nil {
		defer func() {
			log.Print(`Connexion ended`)
			ws.Close()
		}()
	}
	if err != nil {
		log.Printf(`Error while dialing to websocket %s: %v`, url, err)
		return
	}

	ws.WriteMessage(websocket.TextMessage, []byte(secretKey))
	log.Print(`Connection established`)

	go func() {
		time.Sleep(pingDelay)
		ws.WriteMessage(websocket.PingMessage, nil)
	}()

	for {
		messageType, p, err := ws.ReadMessage()
		if messageType == websocket.CloseMessage {
			return
		}
		if err != nil {
			log.Printf(`Error while reading from websocket: %v`, err)
			return
		}

		if messageType == websocket.TextMessage {
			log.Printf(`Received: %s`, p)

			if string(p) == `status` {
				lights, err := listLights(bridgeURL)
				if err != nil {
					err = fmt.Errorf(`Error while listing lights: %v`, err)
					log.Print(err)
					ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
				}

				lightsJSON, err := json.Marshal(lights)
				if err != nil {
					err = fmt.Errorf(`Error while marshalling lights: %v`, err)
					log.Print(err)
					ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
				}

				ws.WriteMessage(websocket.TextMessage, lightsJSON)
			}

			if state, ok := states[string(p)]; ok {
				updateAllState(bridgeURL, state)
			}
		}
	}
}

func main() {
	bridgeIP := flag.String(`bridgeIP`, ``, `IP of Hue Bridge`)
	username := flag.String(`username`, ``, `Username for Hue Bridge`)
	websocketURL := flag.String(`websocket`, ``, `WebSocket URL`)
	secretKey := flag.String(`secretKey`, ``, `Secret Key`)
	state := flag.String(`state`, ``, `State to render`)
	flag.Parse()

	if *websocketURL != `` {
		connect(*websocketURL, getURL(*bridgeIP, *username), *secretKey)
	} else if *state != `` {
		updateAllState(getURL(*bridgeIP, *username), states[*state])
	}
}
