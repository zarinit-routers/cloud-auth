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
	Name         string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Roles        Roles  `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

type Role struct {
	*BaseModel
	Name string `gorm:"many2many:user_roles;" json:"name"`
}

type Roles []Role

func (r Roles) Contains(role string) bool {
	for _, r := range r {
		if r.Name == role {
			return true
		}
	}
	return false
}

func (r Roles) ToSlice() []string {
	var roles []string
	for _, r := range r {
		roles = append(roles, r.Name)
	}
	return roles
}

func (u *User) IsAdmin() bool {
	return u.Roles.Contains("admin")
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
