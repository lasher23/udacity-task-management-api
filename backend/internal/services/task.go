package services

import (
	"task-manager/backend/internal/models"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type TaskService interface {
	CreateTask(db *gorm.DB, task *models.Task) (*models.Task, error)
	UpdateTask(db *gorm.DB, taskID uuid.UUID, updates *models.Task) (*models.Task, error)
	DeleteTask(db *gorm.DB, taskID uuid.UUID) error
	GetTaskByID(db *gorm.DB, taskID uuid.UUID) (*models.Task, error)
	GetTasksByUser(db *gorm.DB, userID uuid.UUID) ([]models.Task, error)
	GetTasks(db *gorm.DB) ([]models.Task, error)
}

type TaskServiceImpl struct {
}

func NewTaskService() *TaskServiceImpl {
	return &TaskServiceImpl{}
}

func (s *TaskServiceImpl) CreateTask(db *gorm.DB, task *models.Task) (*models.Task, error) {
	task.ID = uuid.Must(uuid.NewV7())
	if err := db.Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskServiceImpl) UpdateTask(db *gorm.DB, taskID uuid.UUID, updates *models.Task) (*models.Task, error) {
	var task models.Task
	if err := db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	if updates.Title != "" {
		task.Title = updates.Title
	}
	if updates.Description != "" {
		task.Description = updates.Description
	}
	if updates.Status != "" {
		task.Status = updates.Status
	}
	if updates.DueDate != nil {
		task.DueDate = updates.DueDate
	}
	if err := db.Save(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TaskServiceImpl) DeleteTask(db *gorm.DB, taskID uuid.UUID) error {
	result := db.Delete(&models.Task{}, "id = ?", taskID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (s *TaskServiceImpl) GetTaskByID(db *gorm.DB, taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	if err := db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TaskServiceImpl) GetTasksByUser(db *gorm.DB, userID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	if err := db.Where("user_id = ?", userID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskServiceImpl) GetTasks(db *gorm.DB) ([]models.Task, error) {
	var tasks []models.Task
	if err := db.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}
