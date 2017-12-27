package wemo

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
)

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`webhookKey`: flag.String(tools.ToCamel(prefix+`WebHook`), ``, `[wemo] WebHook Key`),
	}
}

// Handler create Handler from Flags' config
func Handler(config map[string]*string) http.Handler {
	webHookKey := *config[`webhookKey`]

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rawData []byte
		var err error

		if r.URL.Path == `/on` {
			rawData, err = httputils.GetBody(`https://maker.ifttt.com/trigger/wemo_plug_on/with/key/`+webHookKey, nil)
		} else if r.URL.Path == `/off` {
			rawData, err = httputils.GetBody(`https://maker.ifttt.com/trigger/wemo_plug_off/with/key/`+webHookKey, nil)
		} else {
			httputils.NotFound(w)
			return
		}

		if err != nil || strings.HasPrefix(string(rawData), `<!DOCTYPE`) {
			log.Printf(`Error while querying IFTTT WebHook: %v`, err)
			http.Redirect(w, r, fmt.Sprintf(`/?message_level=%s&message_content=%s`, `error`, `Error while requesting WeMo`), http.StatusFound)
		} else {
			http.Redirect(w, r, fmt.Sprintf(`/?message_level=%s&message_content=%s`, `success`, rawData), http.StatusFound)
		}
	})
}
