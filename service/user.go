package service

import (
	"context"
	"errors"
	"time"

	"github.com/snykk/kanban-app/entity"
	"github.com/snykk/kanban-app/repository"
	"github.com/snykk/kanban-app/utils"
)

type UserService interface {
	Login(ctx context.Context, user *entity.User) (id int, err error)
	Register(ctx context.Context, user *entity.User) (entity.User, error)

	GetUserById(ctx context.Context, id int) (entity.User, error)
	Delete(ctx context.Context, id int) error
}

type userService struct {
	userRepository repository.UserRepository
	categoryRepo   repository.CategoryRepository
}

func NewUserService(userRepository repository.UserRepository, categoryRepo repository.CategoryRepository) UserService {
	return &userService{userRepository, categoryRepo}
}

func (s *userService) GetUserById(ctx context.Context, id int) (entity.User, error) {
	return s.userRepository.GetUserByID(ctx, id)
}

func (s *userService) Login(ctx context.Context, user *entity.User) (id int, err error) {
	//check email and password

	dbUser, err := s.userRepository.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return 0, err
	}

	if dbUser.Email == "" || dbUser.ID == 0 {
		return 0, errors.New("user not found")
	}

	if !utils.ValidateHash(user.Password, dbUser.Password) {
		return 0, errors.New("wrong email or password")
	}

	return dbUser.ID, nil
}

func (s *userService) Register(ctx context.Context, user *entity.User) (entity.User, error) {
	dbUser, err := s.userRepository.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return *user, err
	}

	if dbUser.Email != "" || dbUser.ID != 0 {
		return *user, errors.New("email already exists")
	}

	user.Password, err = utils.GenerateHash(user.Password)
	if err != nil {
		return *user, err
	}

	user.CreatedAt = time.Now()

	newUser, err := s.userRepository.CreateUser(ctx, *user)
	if err != nil {
		return *user, err
	}

	// create 4 category
	// Todo, In Progress, Done, Backlog
	categories := []entity.Category{
		{Type: "Todo", UserID: newUser.ID, CreatedAt: time.Now()},
		{Type: "In Progress", UserID: newUser.ID, CreatedAt: time.Now()},
		{Type: "Done", UserID: newUser.ID, CreatedAt: time.Now()},
		{Type: "Backlog", UserID: newUser.ID, CreatedAt: time.Now()},
	}

	err = s.categoryRepo.StoreManyCategory(ctx, categories)
	if err != nil {
		return *user, err
	}

	return newUser, nil
}

func (s *userService) Delete(ctx context.Context, id int) error {
	return s.userRepository.DeleteUser(ctx, id)
}
