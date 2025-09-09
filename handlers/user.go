package handlers

import (
	"net/http"

	"auth-service/database"
	"auth-service/models"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zarinit-routers/middleware/auth"
)

// GetUsers возвращает список всех пользователей
func GetUsers(c *gin.Context) {
	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Скрываем пароли
	for i := range users {
		users[i].PasswordHash = ""
	}

	c.JSON(http.StatusOK, users)
}

// CreateUser создает нового пользователя
func CreateUser(c *gin.Context) {
	var userData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли пользователь с таким email
	var existingUser models.User
	if err := database.DB.Where("email = ?", userData.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists"})
		return
	}

	user := models.User{
		Username: userData.Username,
		Email:    userData.Email,
	}

	if err := user.SetPassword(userData.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
		return
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Скрываем пароль
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, user)
}

func GetSelf(c *gin.Context) {
	// user, ok := c.Get("user")

	data, err := auth.GetUser(c)
	if err != nil {
		log.Error("Failed to fetch user", "error", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	id, err := uuid.Parse(data.UserId)
	if err != nil {
		log.Error("Failed to parse user ID", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	user, err := getUserByID(id)
	if err != nil {
		log.Error("Failed to get user", "id", id, "error", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user.Hide()})
}

func getUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUser возвращает пользователя по ID
func GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	user, err := getUserByID(id)
	if err != nil {
		log.Error("Failed to get user", "id", id, "error", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user.Hide()})
}

// UpdateUser обновляет пользователя
func UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var updateData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := getUserByID(id)
	if err != nil {
		log.Error("Failed to get user", "id", id, "error", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	authData, err := auth.GetUser(c)
	if err != nil {
		log.Error("Failed get auth data", "error", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userHasRights := authData.UserId == user.ID.String() || authData.IsAdmin()
	if !userHasRights {
		log.Error("Update operation not allowed", "authUserID", authData.UserId, "userID", user.ID.String())
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update this user"})
		return
	}

	user.Username = updateData.Username
	user.Email = updateData.Email

	if updateData.Password != "" {
		if err := user.SetPassword(updateData.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
			return
		}
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user.Hide())
}

// DeleteUser удаляет пользователя
func DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	authData, err := auth.GetUser(c)
	if err != nil {
		log.Error("Failed get auth data", "error", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if !authData.IsAdmin() {
		log.Error("Delete operation not allowed for not admins", "authUserID", authData.UserId, "userID", id)
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this user"})
		return
	}
	if id.String() == authData.UserId {
		log.Error("Delete operation not allowed for self", "userID", authData.UserId)
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete yourself"})
		return
	}

	if err := database.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
