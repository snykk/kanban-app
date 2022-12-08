package service

import (
	"context"

	"github.com/snykk/kanban-app/entity"
	"github.com/snykk/kanban-app/repository"
)

type TaskService interface {
	GetTasks(ctx context.Context, id int) ([]entity.Task, error)
	GetTaskByID(ctx context.Context, id int) (entity.Task, error)
	StoreTask(ctx context.Context, task *entity.Task) (entity.Task, error)
	UpdateTask(ctx context.Context, task *entity.Task) (entity.Task, error)
	DeleteTask(ctx context.Context, id int) error
}

type taskService struct {
	taskRepo     repository.TaskRepository
	categoryRepo repository.CategoryRepository
}

func NewTaskService(taskRepo repository.TaskRepository, categoryRepo repository.CategoryRepository) TaskService {
	return &taskService{taskRepo, categoryRepo}
}

func (s *taskService) GetTasks(ctx context.Context, id int) ([]entity.Task, error) {
	return s.taskRepo.GetTasks(ctx, id)
}

func (s *taskService) StoreTask(ctx context.Context, task *entity.Task) (entity.Task, error) {
	_, err := s.taskRepo.StoreTask(ctx, task)
	if err != nil {
		return entity.Task{}, err
	}
	return *task, nil
}

func (s *taskService) GetTaskByID(ctx context.Context, id int) (entity.Task, error) {
	return s.taskRepo.GetTaskByID(ctx, id)
}

func (s *taskService) UpdateTask(ctx context.Context, task *entity.Task) (entity.Task, error) {
	if task.CategoryID != 0 {
		cat, err := s.categoryRepo.GetCategoryByID(ctx, task.CategoryID)
		if err != nil {
			return entity.Task{}, err
		}

		if cat.ID == 0 || cat.Type == "" || cat.UserID != task.UserID {
			return entity.Task{}, err
		}
	}

	err := s.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return entity.Task{}, err
	}
	return *task, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id int) error {
	return s.taskRepo.DeleteTask(ctx, id)
}
