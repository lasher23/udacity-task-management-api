package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	ID          uuid.UUID  `json:"id" gorm:"primaryKey"`
	UserID      uuid.UUID  `json:"user_id" gorm:"not null"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description"`
	Status      string     `json:"status" gorm:"default:pending"`
	Priority    string     `json:"priority" gorm:"default:low"`
	DueDate     *time.Time `json:"due_date"`
}
