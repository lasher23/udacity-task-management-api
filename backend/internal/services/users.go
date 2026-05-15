package services

import (
	"task-manager/backend/internal/models"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	GetUserProfile(db *gorm.DB, userID uuid.UUID) (models.User, error)
	GetUsers(db *gorm.DB) ([]models.User, error)
	DeleteUser(db *gorm.DB, userId uuid.UUID) error
}

type UserServiceImpl struct {
}

func NewUserService() *UserServiceImpl {
	return &UserServiceImpl{}
}

// GetUserProfile fetches a user by ID using a parameterized query (prevents SQL injection).
func (s *UserServiceImpl) GetUserProfile(db *gorm.DB, userID uuid.UUID) (models.User, error) {
	var user models.User

	result := db.Preload("Roles.Permissions").Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return models.User{}, result.Error
	}

	return user, nil
}

func (s *UserServiceImpl) GetUsers(db *gorm.DB) ([]models.User, error) {
	var user []models.User

	result := db.Find(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func (s *UserServiceImpl) DeleteUser(db *gorm.DB, userId uuid.UUID) error {
	result := db.Delete(&models.User{}, "id = ?", userId)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
