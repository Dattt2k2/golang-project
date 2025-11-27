package repository

import (
	"fmt"
	"time"
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
	GetAllUsers(limit, offset int, userType string, status string) (models.PaginatedUsers, error)
	GetUserByID(id uuid.UUID, userType string) (*models.User, error)
	UpdateUserStatus(id uuid.UUID, userType string) error
	AdminDeleteUser(id uuid.UUID, adminType string) error
	GetUserStatistics(month int, year int) (int64, int64, error)
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

func (r *userRepository) GetAllUsers(limit, offset int, userType string, status string) (models.PaginatedUsers, error) {
	if userType == "" {
        return models.PaginatedUsers{}, fmt.Errorf("user Type is nil")
    } else if userType != "ADMIN" {
        return models.PaginatedUsers{}, fmt.Errorf("user Type is invalid")
    }
    var users []models.User
    var total int64
	var isDisabled *bool
	if status == "active" {
		val := false
		isDisabled = &val
	} else if status == "inactive" {
		val := true
		isDisabled = &val
	}
   	countQuery := r.db.Model(&models.User{})
    if isDisabled != nil {
        countQuery = countQuery.Where("is_disabled = ?", *isDisabled)  
    }
    if err := countQuery.Count(&total).Error; err != nil {
        return models.PaginatedUsers{}, err
    }

    query := r.db.Limit(limit).Offset(offset)
    if isDisabled != nil {
        query = query.Where("is_disabled = ?", *isDisabled)  
    }
    if err := query.Find(&users).Error; err != nil {  
        return models.PaginatedUsers{}, err
    }

    hasPrev := offset > 0
    hasNext := int64(offset+limit) < total

    return models.PaginatedUsers{
        Users:   users,
        Total:   total,
        Limit:   limit,
        Offset:  offset,
        HasPrev: hasPrev,
        HasNext: hasNext,
    }, nil
}

func (r *userRepository) GetUserByID(id uuid.UUID, userType string) (*models.User, error) {
	if userType != "ADMIN" {
		return nil, fmt.Errorf("Only ADMIN users can get user by ID")
	}
	var user models.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateUserStatus(id uuid.UUID, userType string) error {
	if userType != "ADMIN" {
		return fmt.Errorf("Only ADMIN users can update user status")
	}
	 var isDisabled bool
    err := r.db.Model(&models.User{}).Where("id = ?", id).Select("is_disabled").Scan(&isDisabled).Error
    if err != nil {
        return err
    }
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("is_disabled", !isDisabled).Error
}

func (r *userRepository) AdminDeleteUser(id uuid.UUID, adminType string) error {
	if adminType != "ADMIN" {
		return fmt.Errorf("Only ADMIN users can delete users")
	}
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) GetUserStatistics(month int, year int) (int64, int64, error) {
	var totalUsers int64
	var previousTotalUsers int64

	buildRange := func(m, y int) (time.Time, time.Time) {
		if m <= 0 || y <= 0 {
			return time.Time{}, time.Time{}
		}
		start := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)
		return start, end
	}

	start, end := buildRange(month, year)
	q := r.db.Model(&models.User{}).Where("users.is_disabled = ?", false)
	if !start.IsZero() && !end.IsZero() {
		q = q.Where("created_at >= ? AND created_at < ?", start, end)
	}
	if err := q.Count(&totalUsers).Error; err != nil {
		return 0, 0, err
	}

	prevStart, prevEnd := buildRange(month-1, year)
	prevQ := r.db.Model(&models.User{}).Where("users.is_disabled = ?", false)
	if !prevStart.IsZero() && !prevEnd.IsZero() {
		prevQ = prevQ.Where("created_at >= ? AND created_at < ?", prevStart, prevEnd)
	}
	if err := prevQ.Count(&previousTotalUsers).Error; err != nil {
		return 0, 0, err
	}

	return totalUsers, previousTotalUsers, nil
}