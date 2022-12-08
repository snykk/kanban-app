package config

import "os"

var (
	// BaseURL is the base url of the server
	BaseURL = os.Getenv("BASE_URL")
)

func SetUrl(url string) string {
	if BaseURL == "" {
		BaseURL = "http://localhost:8080"
	}

	return BaseURL + url
}
