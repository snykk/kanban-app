package repository

import (
	"fmt"

	"github.com/snykk/kanban-app/config"
	"github.com/snykk/kanban-app/entity"

	_ "github.com/jackc/pgx/v4/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

const ENV_PRODUCTION = "production"
const ENV_DEVELOPMENT = "development"

func ConnectDB() error {
	var dsn string
	if config.AppConfig.Environment == ENV_DEVELOPMENT {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s", config.AppConfig.DBUsername, config.AppConfig.DBPassword, config.AppConfig.DBHost, config.AppConfig.DBPort, config.AppConfig.DBDatabase)
	} else if config.AppConfig.Environment == ENV_PRODUCTION {
		dsn = config.AppConfig.DBDsn
	}

	// connect using gorm pgx
	conn, err := gorm.Open(postgres.New(postgres.Config{
		DriverName: "pgx",
		DSN:        dsn,
	}), &gorm.Config{})
	if err != nil {
		return err
	}

	conn.AutoMigrate(entity.User{}, entity.Category{}, entity.Task{})
	SetupDBConnection(conn)

	return nil
}

func SetupDBConnection(DB *gorm.DB) {
	db = DB
}

func GetDBConnection() *gorm.DB {
	return db
}
