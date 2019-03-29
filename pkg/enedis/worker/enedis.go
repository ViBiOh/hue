package worker

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/enedis"
)

const (
	loginURL   = `https://espace-client-connexion.enedis.fr/auth/UI/Login`
	consumeURL = `https://espace-client-particuliers.enedis.fr/group/espace-particuliers/suivi-de-consommation?`

	frenchDateFormat = `02/01/2006`
)

// Config of package
type Config struct {
	email    *string
	password *string
}

// App of package
type App struct {
	email    string
	password string
	cookie   string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		email:    fs.String(tools.ToCamel(fmt.Sprintf(`%sEmail`, prefix)), ``, `[enedis] Email`),
		password: fs.String(tools.ToCamel(fmt.Sprintf(`%sPassword`, prefix)), ``, `[enedis] Password`),
	}
}

// New creates new App from Config
func New(config Config) *App {
	return &App{
		email:    strings.TrimSpace(*config.email),
		password: strings.TrimSpace(*config.password),
	}
}

func (a *App) isEnabled() bool {
	return a.email != `` && a.password != ``
}

func (a *App) appendSessionCookie(headers http.Header) {
	for _, cookie := range headers[`Set-Cookie`] {
		if strings.HasPrefix(cookie, `JSESSIONID`) {
			a.cookie = fmt.Sprintf(`%s; %s`, a.cookie, getCookieValue(cookie))
		}
	}
}

// Login triggers login
func (a *App) Login() error {
	if !a.isEnabled() {
		return nil
	}

	values := url.Values{}
	values.Add(`IDToken1`, a.email)
	values.Add(`IDToken2`, a.password)
	values.Add(`SunQueryParamsString`, `cmVhbG09cGFydGljdWxpZXJz`)
	values.Add(`encoded`, `true`)
	values.Add(`gx_charset`, `UTF-8`)

	ctx := context.Background()
	_, _, headers, err := request.PostForm(ctx, loginURL, values, nil)
	if err != nil {
		return err
	}

	authCookies := strings.Builder{}
	for _, cookie := range headers[`Set-Cookie`] {
		if !strings.Contains(cookie, `Domain=.enedis.fr`) {
			continue
		}

		if strings.Contains(cookie, `Expires=Thu, 01-Jan-1970 00:00:10 GMT`) {
			continue
		}

		if authCookies.Len() != 0 {
			safeWrite(&authCookies, `;`)
		}
		safeWrite(&authCookies, getCookieValue(cookie))
	}

	a.cookie = authCookies.String()

	return nil
}

// GetData retrieve data
func (a *App) GetData(first bool) (*enedis.Consumption, error) {
	if !a.isEnabled() {
		return nil, nil
	}

	header := http.Header{}
	header.Set(`Cookie`, a.cookie)

	params := url.Values{}
	params.Add(`p_p_id`, `lincspartdisplaycdc_WAR_lincspartcdcportlet`)
	params.Add(`p_p_lifecycle`, `2`)
	params.Add(`p_p_state`, `normal`)
	params.Add(`p_p_mode`, `view`)
	params.Add(`p_p_resource_id`, `urlCdcJour`)
	params.Add(`p_p_cacheability`, `cacheLevelPage`)
	params.Add(`p_p_col_id`, `column-1`)
	params.Add(`p_p_col_count`, `2`)

	endDate := time.Now().AddDate(0, 0, -1).Format(frenchDateFormat)
	startDate := time.Now().AddDate(0, 0, -31).Format(frenchDateFormat)

	values := url.Values{}
	params.Add(`_lincspartdisplaycdc_WAR_lincspartcdcportlet_dateDebut`, startDate)
	params.Add(`_lincspartdisplaycdc_WAR_lincspartcdcportlet_dateFin`, endDate)

	ctx := context.Background()
	body, status, headers, err := request.PostForm(ctx, fmt.Sprintf(`%s%s`, consumeURL, params.Encode()), values, header)
	if err != nil || status == http.StatusFound {
		if first {
			a.appendSessionCookie(headers)
			return a.GetData(false)
		}

		return nil, err
	}

	payload, err := request.ReadBody(body)
	if err != nil {
		return nil, err
	}

	var response enedis.Consumption
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, errors.WithStack(err)
	}

	return &response, nil
}

func safeWrite(w *strings.Builder, content string) {
	if _, err := w.WriteString(content); err != nil {
		logger.Error(`%+v`, errors.WithStack(err))
	}
}

func getCookieValue(cookie string) string {
	return strings.SplitN(cookie, `;`, 2)[0]
}
