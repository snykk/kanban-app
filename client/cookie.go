package client

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/snykk/kanban-app/config"
)

func GetClientWithCookie(userID string, cookies ...*http.Cookie) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	cookies = append(cookies, &http.Cookie{
		Name:  "user_id",
		Value: userID,
	})
	data := strings.Split(config.AppConfig.BaseURL, "://")

	jar.SetCookies(&url.URL{
		Scheme: data[0],
		Host:   data[1],
	}, cookies)

	c := &http.Client{
		Jar: jar,
	}

	return c, nil
}
