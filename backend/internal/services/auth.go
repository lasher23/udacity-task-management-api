package services

import (
	"errors"
	"task-manager/backend/internal/models"
	"task-manager/backend/internal/utils"
	"time"

	"github.com/gofrs/uuid"
	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	LoginUser(db *gorm.DB, username, password string) (*models.User, error)
	GenerateToken(db *gorm.DB, userID uuid.UUID) (string, string, error)
	RefreshToken(db *gorm.DB, refreshToken string) (string, string, error)
}

type AuthServiceImpl struct {
}

func NewAuthService() *AuthServiceImpl {
	return &AuthServiceImpl{}
}

func VerifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func (s *AuthServiceImpl) LoginUser(db *gorm.DB, username, password string) (*models.User, error) {
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	if !VerifyPassword(user.Password, password) {
		return nil, errors.New("invalid username or password")
	}

	return &user, nil
}

func (s *AuthServiceImpl) GenerateToken(db *gorm.DB, userID uuid.UUID) (string, string, error) {
	var user models.User
	if err := db.Preload("Roles.Permissions").Where("id = ?", userID).First(&user).Error; err != nil {
		return "", "", err
	}

	exp := time.Now().Add(time.Hour)

	roleNames := make([]string, len(user.Roles))
	permSet := make(map[string]bool)
	for i, r := range user.Roles {
		roleNames[i] = r.Name
		for _, p := range r.Permissions {
			permSet[p.Resource+":"+p.Action] = true
		}
	}
	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":         exp.Unix(),
		"iss":         "task-manager",
		"sub":         userID.String(),
		"user_id":     userID.String(),
		"roles":       roleNames,
		"permissions": permissions,
	})

	signingKey := []byte(utils.GetEnv("JWT_SECRET", ""))
	tokenString, err := accessToken.SignedString(signingKey)
	if err != nil {
		return "", "", err
	}
	token := models.Token{
		ID:           uuid.Must(uuid.NewV7()),
		UserId:       userID,
		RefreshToken: uuid.Must(uuid.NewV7()),
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	if err := db.Create(&token).Error; err != nil {
		return "", "", err
	}
	return tokenString, token.RefreshToken.String(), nil
}
func (s *AuthServiceImpl) RefreshToken(db *gorm.DB, tokenString string) (string, string, error) {
	var token models.Token
	if err := db.Where("refresh_token = ?", tokenString).First(&token).Error; err != nil {
		return "", "", err
	}

	if err := db.Delete(&token).Error; err != nil {
		return "", "", err
	}

	return s.GenerateToken(db, token.UserId)
}
