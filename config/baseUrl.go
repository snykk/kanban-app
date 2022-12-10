package config

func SetUrl(url string) string {
	return AppConfig.BaseURL + url
}
