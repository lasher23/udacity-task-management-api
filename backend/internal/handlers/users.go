package handlers

import (
	"net/http"
	"task-manager/backend/internal/models"
	"task-manager/backend/internal/services"
	"task-manager/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type UserHandler struct {
	db          *gorm.DB
	userService services.UserService
}

func NewUserHandler(db *gorm.DB, userService services.UserService) *UserHandler {
	return &UserHandler{db: db, userService: userService}
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userID, err := utils.ExtractUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	user, err := h.userService.GetUserProfile(h.db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, models.NewUserProfile(user))
}

func (h *UserHandler) GetUserProfileByUserId(c *gin.Context) {
	userID, err := uuid.FromString(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	user, err := h.userService.GetUserProfile(h.db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, models.NewUserProfile(user))
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetUsers(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	profiles := make([]models.UserProfile, len(users))
	for i, u := range users {
		profiles[i] = models.NewUserProfile(u)
	}
	c.JSON(http.StatusOK, profiles)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID, err := uuid.FromString(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.userService.DeleteUser(h.db, userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
