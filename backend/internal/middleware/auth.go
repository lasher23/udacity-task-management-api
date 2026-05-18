package middleware

import (
	"fmt"
	"log"
	"net/http"

	"task-manager/backend/internal/models"
	"task-manager/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

func Authenticate(db *gorm.DB) gin.HandlerFunc {
	return Authorize(db, "", "", false)
}

func Authorize(db *gorm.DB, resource, action string, withOwnership bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.ExtractUserIDFromContext(c)
		if err != nil {
			log.Printf("[AUTH] Unauthorized request to %s: %v", c.FullPath(), err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Set("user_id", userID)

		var user models.User
		if err := db.Preload("Roles.Permissions").Where("id = ?", userID).First(&user).Error; err != nil {
			log.Printf("[AUTH] User not found in DB: %v", userID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if resource == "" {
			c.Next()
			return
		}

		isOwner := false
		if withOwnership {
			if taskIDStr := c.Param("id"); taskIDStr != "" {
				if taskID, err := uuid.FromString(taskIDStr); err == nil {
					var task models.Task
					if db.Where("id = ?", taskID).First(&task).Error == nil && task.UserID == userID {
						isOwner = true
					}
				}
			}
			if c.Param("user_id") == userID.String() {
				isOwner = true
			}

			// Letting admin role bypass ownership
		outer:
			for _, role := range user.Roles {
				if role.Name == "admin" {
					isOwner = true
					break outer
				}
			}
		}

		required := fmt.Sprintf("%s:%s", resource, action)
		hasPermission := false
	role:
		for _, role := range user.Roles {
			for _, p := range role.Permissions {
				if p.Resource+":"+p.Action == required {
					hasPermission = true
					break role
				}
			}
		}

		if (!withOwnership || isOwner) && hasPermission {
			c.Next()
			return
		}

		log.Printf("[AUTH] Access denied for user %v on %s (required=%s, hasPermission=%v, isOwner=%v)", userID, c.FullPath(), required, hasPermission, isOwner)
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		c.Abort()
	}
}
