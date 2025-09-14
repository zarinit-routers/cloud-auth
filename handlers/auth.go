package handlers

import (
	"net/http"
	"time"

	"auth-service/database"
	"auth-service/models"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	"github.com/zarinit-routers/middleware/auth"
)

// Login обрабатывает аутентификацию пользователя
func Login(c *gin.Context) {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		log.Error("Failed bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", loginData.Email).Preload("Roles").First(&user).Error; err != nil {
		log.Error("Failed get user from database", "error", err, "email", loginData.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.CheckPassword(loginData.Password) {
		log.Error("Passwords don't match", "email", loginData.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Создаем claims

	organization, err := getOrganizationId(user)
	if err != nil {
		log.Error("Failed get organization", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization"})
		return
	}

	claims := jwt.MapClaims{
		auth.KeyUserID:         user.ID,
		auth.KeyRoles:          user.Roles.ToSlice(),
		auth.KeyOrganizationID: organization.String(),

		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// Используем единый источник для секретного ключа
	secretKey := []byte(viper.GetString("jwt-security-key"))

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		log.Error("Failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   tokenString,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Name,
			"email":    user.Email,
			"roles":    user.Roles.ToSlice(),
		},
	})
}

// UpdateProfile обновляет профиль пользователя
func UpdateProfile(c *gin.Context) {
	userFromContext, _ := c.Get("user")
	userModel := userFromContext.(models.User)

	var updateData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем поля
	userModel.Name = updateData.Username
	userModel.Email = updateData.Email

	if updateData.Password != "" {
		if err := userModel.SetPassword(updateData.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
			return
		}
	}

	if err := database.DB.Save(&userModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":       userModel.ID,
			"username": userModel.Name,
			"email":    userModel.Email,
			"roles":    userModel.Roles.ToSlice(),
		},
	})
}
