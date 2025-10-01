package services

import (
	"errors"
	"user-service/internal/events"
	"user-service/internal/models"
	"user-service/internal/repository"

	"github.com/google/uuid"
)

type UserService struct {
	repo      repository.UserRepository
	publisher events.EventPublisher
}

func NewUserService(repo repository.UserRepository, publisher events.EventPublisher) *UserService {
	return &UserService{repo: repo, publisher: publisher}
}

func (s *UserService) CreateUserService(user *models.User) error {
	if user.Email == nil || *user.Email == "" {
		return errors.New("email is required")
	}
	if err := s.repo.SaveUser(user); err != nil {
		return err
	}

	// Publish created event (without password)
	if s.publisher != nil {
		payload := map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"phone":      user.Phone,
			"user_type":  user.UserType,
		}
		_ = s.publisher.Publish("user.created", payload)
	}
	return nil
}

func (s *UserService) GetUserService(id uuid.UUID) (*models.User, error) {
	return s.repo.FindUserByID(id)
}

func (s *UserService) UpdateUserService(user *models.User) error {
	if user.ID == uuid.Nil {
		return errors.New("user ID is required")
	}
	return s.repo.UpdateUser(user)
}

func (s *UserService) DeleteUserService(id uuid.UUID) error {
	return s.repo.DeleteUser(id)
}
