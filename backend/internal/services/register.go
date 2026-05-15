package services

import (
	"task-manager/backend/internal/models"

	"gorm.io/gorm"
)

type RegisterService interface {
	RegisterUser(db *gorm.DB, user models.User) error
}

type RegisterServiceImpl struct{}

func NewRegisterService() *RegisterServiceImpl {
	return &RegisterServiceImpl{}
}

func (s *RegisterServiceImpl) RegisterUser(db *gorm.DB, user models.User) error {
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	var role models.Role
	if err := db.Where("name = ?", "user").First(&role).Error; err != nil {
		return err
	}
	return db.Model(&user).Association("Roles").Append(&role)
}
