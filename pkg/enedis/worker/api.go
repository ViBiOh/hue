package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/iot/pkg/enedis"
)

// Login triggers login
func (a *App) Login() error {
	if !a.Enabled() {
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
func (a *App) GetData(ctx context.Context, first bool) (*enedis.Consumption, error) {
	if !a.Enabled() {
		return nil, nil
	}

	header := http.Header{}
	header.Set(`Cookie`, a.cookie)

	params := url.Values{}
	params.Add(`p_p_id`, `lincspartdisplaycdc_WAR_lincspartcdcportlet`)
	params.Add(`p_p_lifecycle`, `2`)
	params.Add(`p_p_state`, `normal`)
	params.Add(`p_p_mode`, `view`)
	params.Add(`p_p_resource_id`, `urlCdcHeure`)
	params.Add(`p_p_cacheability`, `cacheLevelPage`)
	params.Add(`p_p_col_id`, `column-1`)
	params.Add(`p_p_col_count`, `2`)

	date := time.Now().AddDate(0, 0, -1).Format(frenchDateFormat)

	values := url.Values{}
	params.Add(`_lincspartdisplaycdc_WAR_lincspartcdcportlet_dateDebut`, date)
	params.Add(`_lincspartdisplaycdc_WAR_lincspartcdcportlet_dateFin`, date)

	body, status, headers, err := request.PostForm(ctx, fmt.Sprintf(`%s%s`, consumeURL, params.Encode()), values, header)
	if err != nil || status == http.StatusFound {
		if first {
			a.appendSessionCookie(headers)
			return a.GetData(ctx, false)
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
