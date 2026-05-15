package handlers

import (
	"log"
	"net/http"

	"task-manager/backend/internal/models"
	"task-manager/backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterHandler struct {
	db              *gorm.DB
	registerService services.RegisterService
}

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func NewRegisterHandler(db *gorm.DB, registerService services.RegisterService) *RegisterHandler {
	return &RegisterHandler{db: db, registerService: registerService}
}

func (h *RegisterHandler) Registration(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Unexpected error while hashing password: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while creating user"})
		return
	}

	user := models.User{
		ID:       uuid.Must(uuid.NewV7()),
		Username: input.Username,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	if err := h.registerService.RegisterUser(h.db, user); err != nil {
		log.Println("Unexpected error while creating user: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unexpected error occurred while creating user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})
}
