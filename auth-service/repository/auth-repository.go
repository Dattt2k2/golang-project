package repository

import (
	"context"
	"fmt"

	"auth-service/database"
	"auth-service/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetAllUsers(ctx context.Context, offset int, limit int) ([]models.User, error)
	UpdatePassword(ctx context.Context, userID string, hashedPass string) error
	GetUserType(ctx context.Context, userID string) (string, error)
	UpdateVerificationStatus(ctx context.Context, email string, isVerified bool) error
	UpdateRole(ctx context.Context, userID, newRole string) error
	UpdateDisabled(ctx context.Context, userID string, isDisabled bool) error
	FindByEmailAny(ctx context.Context, email string) (*models.User, error)
	DeleteUser(ctx context.Context, userID string) error
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepositoryImpl{
		db: database.DB,
	}
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ? AND is_verify = true AND is_disabled = false", email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *models.User) (*models.User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) GetAllUsers(ctx context.Context, startIndex int, recordPerPage int) ([]models.User, error) {

	var users []models.User
	if err := r.db.WithContext(ctx).Offset(startIndex).Limit(recordPerPage).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepositoryImpl) FindByEmailAny(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error
}

func (r *userRepositoryImpl) GetUserType(ctx context.Context, userID string) (string, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Select("user_type").Where("id = ?", userID).First(&user).Error; err != nil {
		return "", err
	}
	return user.UserType, nil
}

func (r *userRepositoryImpl) UpdateVerificationStatus(ctx context.Context, email string, isVerified bool) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("email = ?", email).
		Update("is_verify", isVerified).Error
}

func (r *userRepositoryImpl) UpdateRole(ctx context.Context, userID, newRole string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("user_type", newRole).Error
}

func (r *userRepositoryImpl) UpdateDisabled(ctx context.Context, userID string, isDisabled bool) error {
	res := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_disabled", isDisabled)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("no rows updated for user id=%s", userID)
	}
	return nil
}

func (r *userRepositoryImpl) DeleteUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("id = ?", userID).Delete(&models.User{}).Error
}
