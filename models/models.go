package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type User struct {
	*BaseModel
	Username     string `gorm:"unique;not null" json:"username"`
	Email        string `gorm:"unique;not null" json:"email"`
	PasswordHash string `gorm:"not null" json:"-"`
	Roles        []Role `json:"roles,omitempty"`
}

type Role struct {
	*BaseModel
	Name  string `gorm:"unique;not null" json:"name"`
	Users []User `json:"users,omitempty"`
}

type UserRole struct {
	UserID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null;primary_key" json:"userId"`
	RoleID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null;primary_key" json:"roleId"`
}

func (u *User) IsAdmin() bool {
	for _, role := range u.Roles {
		if role.Name == "admin" {
			return true
		}
	}
	return false
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// Hides password
func (u *User) Hide() *User {
	u.PasswordHash = "***"
	return u
}
