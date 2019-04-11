package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/pkg/errors"
)

func getAuth(email, password, country string) (map[string]string, error) {
	data := url.Values{
		"Email":    []string{email},
		"Password": []string{password},
	}

	loginRequest, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s?country=%s", API, authenticateEndpoint, country), strings.NewReader(data.Encode()))
	loginRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		return nil, errors.WithStack(err)
	}

	body, _, _, err := request.DoAndReadWithClient(nil, unsafeHTTPClient, loginRequest)
	if err != nil {
		return nil, err
	}

	payload, err := request.ReadBody(body)
	if err != nil {
		return nil, err
	}

	var authContent map[string]string
	if err = json.Unmarshal(payload, &authContent); err != nil {
		return nil, errors.WithStack(err)
	}

	return authContent, nil
}
