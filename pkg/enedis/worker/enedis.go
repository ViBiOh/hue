package worker

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
)

const (
	loginURL = `https://espace-client-connexion.enedis.fr/auth/UI/Login`
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
	req, err := request.Form(http.MethodPost, loginURL, values, nil)
	if err != nil {
		return err
	}

	_, _, headers, err := request.DoAndRead(ctx, req)
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

		value := strings.SplitN(cookie, `;`, 2)
		if _, err := authCookies.WriteString(value[0]); err != nil {
			return errors.WithStack(err)
		}
	}

	a.cookie = authCookies.String()

	logger.Info(`%s`, a.cookie)

	return nil
}
