package database

import (
	"auth-service/models"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB инициализирует подключение к базе данных
func InitDB() (*gorm.DB, error) {
	dsn := viper.GetString("db-connection-string")
	if dsn == "" {
		return nil, fmt.Errorf("db-connection-string is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Получаем underlying sql.DB для настройки
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Устанавливаем параметры соединения
	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Проверяем соединение
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	log.Info("Database connection established")
	return db, nil
}

// MigrateModels выполняет миграцию с обработкой ошибок
func MigrateModels(db *gorm.DB) error {

	tables := []any{
		&models.User{},
		&models.Role{},
	}

	if err := db.AutoMigrate(tables...); err != nil {
		// Игнорируем ошибку отсутствия ограничения
		if !strings.Contains(err.Error(), "does not exist") {
			log.Error("Warning during migration", "error", err)
			return fmt.Errorf("failed to migrate models: %w", err)
		}
	}

	log.Info("Models migrated successfully")
	return nil
}
