package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       uuid.UUID `json:"id" gorm:"primaryKey"`
	Username string    `json:"username" gorm:"unique"`
	Email    string    `json:"email" gorm:"unique"`
	Password string    `json:"-"`
	Roles    []Role    `json:"roles" gorm:"many2many:user_roles;"`
}

type UserProfile struct {
	ID uuid.UUID `json:"id"`
	// Frontend is buggy otherwise need id and user_id
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"CreatedAt"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
}

func NewUserProfile(u User) UserProfile {
	roleNames := make([]string, len(u.Roles))
	permSet := make(map[string]bool)
	for i, r := range u.Roles {
		roleNames[i] = r.Name
		for _, p := range r.Permissions {
			permSet[p.Resource+":"+p.Action] = true
		}
	}
	perms := make([]string, 0, len(permSet))
	for p := range permSet {
		perms = append(perms, p)
	}
	return UserProfile{
		ID:          u.ID,
		UserID:      u.ID,
		Username:    u.Username,
		Email:       u.Email,
		CreatedAt:   u.CreatedAt,
		Roles:       roleNames,
		Permissions: perms,
	}
}
