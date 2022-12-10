package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

var AppConfig Config

type Config struct {
	Port        int
	Environment string
	Debug       bool
	BaseURL     string

	DBHost     string
	DBPort     int
	DBDatabase string
	DBUsername string
	DBPassword string
	DBDsn      string

	JWTSecret  string
	JWTExpired int
	JWTIssuer  string

	OTPEmail    string
	OTPPassword string

	REDISHost     string
	REDISPassword string
	REDISExpired  int
}

var ERRORS_EMPTY_ENV = errors.New("required variabel environment is empty")

func InitializeAppConfig() error {
	viper.SetConfigName(".env") // allow directly reading from .env file
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("config")
	viper.AddConfigPath("/")
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	// assign value
	AppConfig.Port = viper.GetInt("PORT")
	AppConfig.Environment = viper.GetString("ENVIRONMENT")
	AppConfig.Debug = viper.GetBool("DEBUG")
	AppConfig.BaseURL = viper.GetString("BASE_URL")

	AppConfig.DBHost = viper.GetString("DB_HOST")
	AppConfig.DBPort = viper.GetInt("DB_PORT")
	AppConfig.DBDatabase = viper.GetString("DB_DATABASE")
	AppConfig.DBUsername = viper.GetString("DB_USERNAME")
	AppConfig.DBPassword = viper.GetString("DB_PASSWORD")
	AppConfig.DBDsn = viper.GetString("DB_DSN")

	// check
	if AppConfig.Port == 0 || AppConfig.Environment == "" || AppConfig.BaseURL == "" {
		return ERRORS_EMPTY_ENV
	}

	switch AppConfig.Environment {
	case "development":
		if AppConfig.DBHost == "" || AppConfig.DBPort == 0 || AppConfig.DBDatabase == "" || AppConfig.DBUsername == "" || AppConfig.DBPassword == "" {
			return ERRORS_EMPTY_ENV
		}
	case "production":
		if AppConfig.DBDsn == "" {
			return ERRORS_EMPTY_ENV
		}
	}

	log.Println("[INIT] configuration loaded")
	return nil
}
