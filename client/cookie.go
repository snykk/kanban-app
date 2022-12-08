package client

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
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

	jar.SetCookies(&url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
	}, cookies)

	c := &http.Client{
		Jar: jar,
	}

	return c, nil
}
