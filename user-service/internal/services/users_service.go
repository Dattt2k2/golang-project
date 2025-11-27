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
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		return err
	}

	err = s.repo.DeleteUser(id)
	if err != nil {
		return err
	}

	if s.publisher != nil {
		payload := map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		}
		_ = s.publisher.Publish("user.deleted", payload)
	}
	return nil
}

func (s *UserService) AddAddressService(address *models.UserAddress) error {
	return s.repo.CreateAddress(address)
}

func (s *UserService) UpdateAddressService(address *models.UserAddress) error {
	return s.repo.UpdateAddress(address)
}

func (s *UserService) GetUserAddressesService(userID uuid.UUID, limit, offset int) ([]models.UserAddress, error) {
	return s.repo.GetAddresses(userID, limit, offset)
}

func (s *UserService) DeleteAddressService(addressID uuid.UUID) error {
	return s.repo.DeleteAddress(addressID)
}

func (s *UserService) ListUsersService(limit, offset int, userType string, status string) (models.PaginatedUsers, error) {

	return s.repo.GetAllUsers(limit, offset, userType, status)
}

func (s *UserService) GetUserByIDService(id uuid.UUID, userType string) (*models.User, error) {
	return s.repo.GetUserByID(id, userType)
}

func (s *UserService) UpdateUserStatusService(id uuid.UUID, userType string) error {
	if err := s.repo.UpdateUserStatus(id, userType); err != nil {
		return err
	}

	// Fetch updated user to include the current status
	updatedUser, err := s.repo.FindUserByID(id)
	if err == nil && s.publisher != nil {
		emailVal := ""
		if updatedUser.Email != nil {
			emailVal = *updatedUser.Email
		}
		payload := map[string]interface{}{
			"id":          updatedUser.ID.String(),
			"email":       emailVal,
			"is_disabled": updatedUser.IsDisabled,
		}
		// publish user.disabled event (or user.status.updated) so downstream services can react
		_ = s.publisher.Publish("user.disabled", payload)
	}
	return nil
}

func (s *UserService) AdminDeleteUserService(id uuid.UUID, adminType string) error {
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		return err
	}
	if user.UserType == nil || *user.UserType == "ADMIN" {
		return errors.New("cannot delete ADMIN users")
	}

	err = s.repo.AdminDeleteUser(id, adminType)
	if err != nil {
		return err
	}

	if s.publisher != nil {
		payload := map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		}
		_ = s.publisher.Publish("user.deleted", payload)
	}
	return nil
}
