package worker

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
)

const (
	loginURL   = `https://espace-client-connexion.enedis.fr/auth/UI/Login`
	consumeURL = `https://espace-client-particuliers.enedis.fr/group/espace-particuliers/suivi-de-consommation?`
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
			if _, err := authCookies.WriteString(`;`); err != nil {
				return errors.WithStack(err)
			}
		}

		if _, err := authCookies.WriteString(strings.SplitN(cookie, `;`, 2)[0]); err != nil {
			return errors.WithStack(err)
		}
	}

	a.cookie = authCookies.String()

	return nil
}

// GetData retrieve data
func (a *App) GetData(first bool) ([]byte, error) {
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

	values := url.Values{}
	params.Add(`_lincspartdisplaycdc_WAR_lincspartcdcportlet_dateDebut`, `24/02/2019`)
	params.Add(`_lincspartdisplaycdc_WAR_lincspartcdcportlet_dateFin`, `26/03/2019`)

	ctx := context.Background()
	body, status, headers, err := request.PostForm(ctx, fmt.Sprintf(`%s%s`, consumeURL, params.Encode()), values, header)
	if err != nil || status == http.StatusFound {
		if first {
			for _, cookie := range headers[`Set-Cookie`] {
				if strings.HasPrefix(cookie, `JSESSIONID`) {
					a.cookie = fmt.Sprintf(`%s; %s`, a.cookie, strings.SplitN(cookie, `;`, 2)[0])
				}
			}

			return a.GetData(false)
		}

		return nil, err
	}

	payload, err := request.ReadBody(body)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
