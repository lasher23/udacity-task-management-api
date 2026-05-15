package handlers

import (
	"errors"
	"log"
	"net/http"

	"task-manager/backend/internal/models"
	"task-manager/backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type TaskHandler struct {
	db          *gorm.DB
	taskService services.TaskService
}

func NewTaskHandler(db *gorm.DB, taskService services.TaskService) *TaskHandler {
	return &TaskHandler{db: db, taskService: taskService}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.taskService.CreateTask(h.db, &task)
	if err != nil {
		handleTaskError(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	taskID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.taskService.UpdateTask(h.db, taskID, &task)
	if err != nil {
		handleTaskError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	if err := h.taskService.DeleteTask(h.db, taskID); err != nil {
		handleTaskError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	taskID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	task, err := h.taskService.GetTaskByID(h.db, taskID)
	if err != nil {
		handleTaskError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) GetTasksByUser(c *gin.Context) {
	userID, err := uuid.FromString(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	tasks, err := h.taskService.GetTasksByUser(h.db, userID)
	if err != nil {
		handleTaskError(c, err)
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	tasks, err := h.taskService.GetTasks(h.db)
	if err != nil {
		handleTaskError(c, err)
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func handleTaskError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "task not found",
		})
	} else {
		log.Println("Unexpected error while processing task request: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to process task request",
		})
	}
}
