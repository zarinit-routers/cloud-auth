package main

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"github.com/zarinit-routers/middleware/auth"

	"auth-service/database"
	"auth-service/handlers"
	"auth-service/middleware"
	"auth-service/models"
)

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
func initDatabase() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}

	if err := database.MigrateModels(db); err != nil {
		log.Fatal("Failed to migrate models", "error", err)
	}
}

func main() {
	initDatabase()
	ensureAdminCreated()

	r := gin.Default()

	r.Use(middleware.CORS())

	// API endpoints
	{
		api := r.Group("/api/auth")
		api.POST("/login", handlers.Login)
		api.GET("/users/me", auth.Middleware(), handlers.GetUser)
		api.GET("/users", auth.Middleware(auth.AdminOnly()), handlers.GetUsers)
		api.GET("/users/:id", auth.Middleware(), handlers.GetUser)
		api.POST("/users/:id", auth.Middleware(), handlers.UpdateUser)
		api.POST("/users", handlers.CreateUser)
		api.DELETE("/users/:id", handlers.DeleteUser)
	}

	// Запуск сервера
	viper.SetDefault("port", 5001)
	port := viper.GetInt("port")

	log.Info("Server starting", "port", port)
	r.Run(fmt.Sprintf(":%d", port))
}

// ensureAdminCreated создает root пользователя если не существует
func ensureAdminCreated() {

	viper.SetDefault("root_user_email", "root@admin.com")
	viper.SetDefault("root_user_password", "admin123")

	rootEmail := viper.GetString("root_user_email")
	rootPassword := viper.GetString("root_user_password")

	log.Info("Создание пользователя рут", "email", rootEmail, "password", rootPassword)

	var adminRole models.Role
	if err := database.DB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		log.Warn("Failed to get admin role, creating new", "error", err)
		adminRole = *createAdminRole()
	}

	var count int64
	database.DB.Model(&models.User{}).Where("email = ?", rootEmail).Count(&count)
	if count != 0 {
		log.Warn("User already exists, skipping", "email", rootEmail)
		return
	}
	user := models.User{
		Username: "root",
		Email:    rootEmail,
		Roles:    []models.Role{adminRole},
	}

	if err := user.SetPassword(rootPassword); err != nil {
		log.Error("Failed to set password for root user", "error", err)
		return
	}

	if err := database.DB.Create(&user).Error; err != nil {
		log.Error("Failed to create root user", "error", err)
	} else {
		log.Info("Root user created successfully")
	}
}

func createAdminRole() *models.Role {
	role := models.Role{
		Name: "admin",
	}

	if err := database.DB.Create(&role).Error; err != nil {
		log.Fatal("Failed to create admin role", "error", err)
	} else {
		log.Info("Admin role created")
	}
	return &role
}
