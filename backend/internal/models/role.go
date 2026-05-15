package models

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Permission struct {
	ID       uuid.UUID `json:"id" gorm:"primaryKey"`
	Resource string    `json:"resource"`
	Action   string    `json:"action"`
}

type Role struct {
	gorm.Model
	ID          string       `gorm:"primaryKey"`
	Name        string       `gorm:"unique"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}
