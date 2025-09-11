package main

import (
	"auth-service/database"
	"flag"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"
)

var config struct {
	operation migrate.MigrationDirection
}

func init() {
	config.operation = migrate.Up
	flag.BoolFunc("down", "", func(_ string) error {
		config.operation = migrate.Down
		return nil
	})
	flag.Parse()
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Warn("Error occurred while loading .env file", "error", err)
	}

	viper.AutomaticEnv()
	viper.SetConfigName("cloud-auth-config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Warn("Error occurred while reading config file:", err)
	}

}

func main() {

	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Error occurred while initializing database", "error", err)
	}

	if config.operation == migrate.Up {
		if err := database.Migrate(db); err != nil {
			log.Fatal("Error occurred while migrating database", "error", err)
		}
		log.Info("Migrated")
	} else {
		if err := database.MigrateDown(db); err != nil {
			log.Fatal("Error occurred while rolling back database", "error", err)
		}
		log.Info("Rolled back")
	}
}
