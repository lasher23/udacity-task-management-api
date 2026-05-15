package handlers

import (
	"log"
	"net/http"
	"task-manager/backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db          *gorm.DB
	authService services.AuthService
}

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(db *gorm.DB, authService services.AuthService) *AuthHandler {
	return &AuthHandler{db: db, authService: authService}
}

func (h *AuthHandler) Token(c *gin.Context) {
	// Implement the login and token generation logic here
	var authReq AuthRequest
	if err := c.ShouldBindJSON(&authReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.authService.LoginUser(h.db, authReq.Username, authReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	accessToken, refreshToken, err := h.authService.GenerateToken(h.db, user.ID)
	if err != nil {
		log.Println("Unexpected error while generating token: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    3600,
	})
}
