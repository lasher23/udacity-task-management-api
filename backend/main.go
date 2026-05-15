package main

import (
	"log"
	"task-manager/backend/internal/handlers"
	"task-manager/backend/internal/middleware"
	"task-manager/backend/internal/repositories"
	"task-manager/backend/internal/services"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	dbCfg := repositories.NewDatabaseConfig()
	db, err := dbCfg.Connect()
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance: ", err)
	}
	defer sqlDB.Close()

	authHandler := handlers.NewAuthHandler(db, services.NewAuthService())

	taskHandler := handlers.NewTaskHandler(db, services.NewTaskService())

	refreshHandler := handlers.NewRefreshHandler(db, services.NewAuthService())

	userHandler := handlers.NewUserHandler(db, services.NewUserService())

	registrationHandler := handlers.NewRegisterHandler(db, services.NewRegisterService())

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://host.docker.internal", "http://localhost:8081"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	v1 := r.Group("/api/v1")
	{
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", registrationHandler.Registration)
			authRoutes.POST("/login", authHandler.Token)
			authRoutes.POST("/refresh", refreshHandler.Refresh)
		}
		taskRoutes := v1.Group("/tasks")
		{
			taskRoutes.POST("", middleware.Authorize(db, "task", "write", false), taskHandler.CreateTask)
			taskRoutes.PUT("/:id", middleware.Authorize(db, "task", "write", true), taskHandler.UpdateTask)
			taskRoutes.DELETE("/:id", middleware.Authorize(db, "task", "delete", false), taskHandler.DeleteTask)
			taskRoutes.GET("/:id", middleware.Authorize(db, "task", "read", true), taskHandler.GetTaskByID)
			taskRoutes.GET("", middleware.Authorize(db, "task", "read", false), taskHandler.GetTasks)
		}
		userRoutes := v1.Group("/users")
		{
			userRoutes.DELETE("/:user_id", middleware.Authorize(db, "user", "delete", false), userHandler.DeleteUser)
			userRoutes.GET("", middleware.Authorize(db, "user", "read", false), userHandler.GetUsers)
			userRoutes.GET("/:user_id/tasks", middleware.Authorize(db, "task", "read", true), taskHandler.GetTasksByUser)
			userRoutes.GET("/profile", middleware.Authenticate(db), userHandler.GetUserProfile)
			userRoutes.GET("/profile/:user_id", middleware.Authorize(db, "user", "read", false), userHandler.GetUserProfileByUserId)
		}
	}
	r.Run(":8080")
}
