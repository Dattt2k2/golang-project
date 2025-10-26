package repository

import (
	"user-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	SaveUser(user *models.User) error
	FindUserByID(id uuid.UUID) (*models.User, error)
	DeleteUser(id uuid.UUID) error
	UpdateUser(user *models.User) error
	CreateAddress(address *models.UserAddress) error
	UpdateAddress(address *models.UserAddress) error
	GetAddresses(userID uuid.UUID, limit, offset int) ([]models.UserAddress, error)
	DeleteAddress(id uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) SaveUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) DeleteUser(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) CreateAddress(address *models.UserAddress) error {
	tx := r.db.Begin()

	if address.IsDefault {
		if err := tx.Model(&models.UserAddress{}).
			Where("user_id = ?", address.UserID).
			Update("is_default", false).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Create(address).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Model(&models.User{}).
			Where("id = ?", address.UserID).
			Update("default_address_id", address.ID).Error; err != nil {
			tx.Rollback()
			return err
		}

	return tx.Commit().Error
}

func (r *userRepository) UpdateAddress(address *models.UserAddress) error {
	tx := r.db.Begin()

	if address.IsDefault {
		if err := tx.Model(&models.UserAddress{}).
			Where("user_id = ?", address.UserID).
			Update("is_default", false).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Model(&models.User{}).
			Where("id = ?", address.UserID).
			Update("default_address_id", address.ID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Save(address).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *userRepository) GetAddresses(userID uuid.UUID, limit, offset int) ([]models.UserAddress, error) {
	var addresses []models.UserAddress
	if err := r.db.Where("user_id = ?", userID).Order("is_default DESC").Limit(limit).Offset(offset).Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *userRepository) DeleteAddress(id uuid.UUID) error {
	tx := r.db.Begin()

	var address models.UserAddress
	if err := tx.First(&address, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&models.UserAddress{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if address.IsDefault {
		var newDefaultAddress models.UserAddress
		if err := tx.Where("user_id = ?", address.UserID).
            Order("created_at DESC").
            First(&newDefaultAddress).Error; err == nil {
            if err := tx.Model(&models.UserAddress{}).
                Where("id = ?", newDefaultAddress.ID).
                Update("is_default", true).Error; err != nil {
                tx.Rollback()
                return err
            }

            if err := tx.Model(&models.User{}).
                Where("id = ?", address.UserID).
                Update("default_address_id", newDefaultAddress.ID).Error; err != nil {
                tx.Rollback()
                return err
            }
        } else {
            if err := tx.Model(&models.User{}).
                Where("id = ?", address.UserID).
                Update("default_address_id", nil).Error; err != nil {
                tx.Rollback()
                return err
            }
        }
	}
	return tx.Commit().Error
}
