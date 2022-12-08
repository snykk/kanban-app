package service

import (
	"context"

	"github.com/snykk/kanban-app/entity"
	"github.com/snykk/kanban-app/repository"
)

type CategoryService interface {
	GetCategories(ctx context.Context, id int) ([]entity.Category, error)
	StoreCategory(ctx context.Context, category *entity.Category) (entity.Category, error)
	GetCategoryByID(ctx context.Context, id int) (entity.Category, error)
	UpdateCategory(ctx context.Context, category *entity.Category) (entity.Category, error)
	DeleteCategory(ctx context.Context, id int) error
	GetCategoriesWithTasks(ctx context.Context, id int) ([]entity.CategoryData, error)
}

type categoryService struct {
	catRepo  repository.CategoryRepository
	taskRepo repository.TaskRepository
}

func NewCategoryService(catRepo repository.CategoryRepository, taskRepo repository.TaskRepository) CategoryService {
	return &categoryService{catRepo, taskRepo}
}

func (s *categoryService) GetCategories(ctx context.Context, id int) ([]entity.Category, error) {
	return s.catRepo.GetCategoriesByUserId(ctx, id)
}

func (s *categoryService) StoreCategory(ctx context.Context, category *entity.Category) (entity.Category, error) {
	_, err := s.catRepo.StoreCategory(ctx, category)
	if err != nil {
		return entity.Category{}, err
	}
	return *category, nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id int) (entity.Category, error) {
	return s.catRepo.GetCategoryByID(ctx, id)
}

func (s *categoryService) UpdateCategory(ctx context.Context, category *entity.Category) (entity.Category, error) {
	err := s.catRepo.UpdateCategory(ctx, category)
	if err != nil {
		return entity.Category{}, err
	}
	return *category, nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id int) error {
	tasks, err := s.taskRepo.GetTasksByCategoryID(ctx, id)
	if err != nil {
		return err
	}

	if len(tasks) > 0 {
		for _, task := range tasks {
			err := s.taskRepo.DeleteTask(ctx, task.ID)
			if err != nil {
				return err
			}
		}
	}

	return s.catRepo.DeleteCategory(ctx, id)
}

func (s *categoryService) GetCategoriesWithTasks(ctx context.Context, id int) ([]entity.CategoryData, error) {
	categories, err := s.catRepo.GetCategoriesByUserId(ctx, id)
	if err != nil {
		return nil, err
	}

	tasks, err := s.taskRepo.GetTasks(ctx, id)
	if err != nil {
		return nil, err
	}

	var categoryData = entity.DataToCategoryData(categories, tasks)
	return categoryData, nil
}
