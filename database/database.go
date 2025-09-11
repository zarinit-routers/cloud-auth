package database

import (
	"fmt"

	"github.com/charmbracelet/log"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB инициализирует подключение к базе данных
func InitDB() (*gorm.DB, error) {
	dsn := viper.GetString("db_connection_string")
	if dsn == "" {
		return nil, fmt.Errorf("db_connection_string is not set")
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

// Migrate выполняет миграцию с обработкой ошибок
func Migrate(db *gorm.DB) error {
	return migrateDB(db, migrate.Up)
}

func MigrateDown(db *gorm.DB) error {
	return migrateDB(db, migrate.Down)
}

func migrateDB(db *gorm.DB, direction migrate.MigrationDirection) error {
	migrations := getMigrationSource()
	sqlDb, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	count, err := migrate.Exec(sqlDb, "postgres", migrations, direction)
	if err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}
	log.Info("Executed migrations", "count", count)
	return nil

}

func getMigrationSource() migrate.MigrationSource {
	return &migrate.FileMigrationSource{
		Dir: "migrations",
	}
}
