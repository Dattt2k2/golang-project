package service

import (
	"context"
	"errors"
	"time"

	"auth-service/helpers"
	"auth-service/models"
	"auth-service/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService interface {
	Register(ctx context.Context, email, password, userType string) (*models.SignUpResponse, error)
	Login(ctx context.Context, credential *models.LoginCredentials) (*models.LoginResponse, error)
	GetAllUsers(ctx context.Context, page int, recordPage int) ([]interface{}, error)
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserType(ctx context.Context, userID string) (string, error)
	ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error
	AdminChangePassword(ctx context.Context, adminID, targetUserID, newPassword string) error
	Logout(ctx context.Context, userID string, deviceId string) error
	LogoutAll(ctx context.Context, userID string) error
}

type authServiceImpl struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authServiceImpl{
		userRepo: userRepo,
	}
}

func (s *authServiceImpl) Register(ctx context.Context, email, password, userType string) (*models.SignUpResponse, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	emailExists, err := helpers.CheckEmailExists(email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		return nil, err
	}

	if userType == "" {
		userType = "USER"
	}

	defaultFirstName := "User"
	defaultLastName := ""
	defaultPhone := ""

	user := &models.User{
		ID:         primitive.NewObjectID(),
		Email:      &email,
		Password:   &hashedPassword,
		First_name: &defaultFirstName,
		Last_name:  &defaultLastName,
		User_type:  &userType,
		Phone:      &defaultPhone,
		Created_at: time.Now(),
		Updated_at: time.Now(),
		IsVerify:   true,
	}
	user.User_id = user.ID.Hex()

	token, refreshToken, err := helpers.GenerateAllToken(email, defaultFirstName, defaultLastName, userType, user.User_id)
	if err != nil {
		return nil, err
	}
	result, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	err = helpers.AddUserToBloomFilter(*user.Email, *user.Phone)
	if err != nil {
		return nil, err
	}

	return &models.SignUpResponse{
		Message:      "User registered successfully",
		User:         result,
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authServiceImpl) Login(ctx context.Context, credential *models.LoginCredentials) (*models.LoginResponse, error) {
	foundUser, err := s.userRepo.FindByEmail(ctx, *credential.Email)
	if err != nil {
		return nil, errors.New("Email or password is incorrect")
	}

	if !helpers.VerifyPassword(*credential.Password, *foundUser.Password) {
		return nil, errors.New("Email or password is incorrect")
	}

	token, refreshToken, err := helpers.GenerateAllToken(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
	if err != nil {
		return nil, errors.New("Error generating token")
	}

	loginResponse := &models.LoginResponse{
		Email:        *foundUser.Email,
		First_name:   *foundUser.First_name,
		Last_name:    *foundUser.Last_name,
		User_type:    *foundUser.User_type,
		User_id:      foundUser.User_id,
		Token:        token,
		RefreshToken: refreshToken,
	}

	return loginResponse, nil
}

func (s *authServiceImpl) GetAllUsers(ctx context.Context, page int, recordPerPage int) ([]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if recordPerPage < 1 {
		recordPerPage = 10
	}

	startIndex := (page - 1) * recordPerPage
	users, err := s.userRepo.GetAllUsers(ctx, startIndex, recordPerPage)
	if err != nil {
		return nil, err
	}

	// Convert []primitive.M to []interface{}
	result := make([]interface{}, len(users))
	for i, v := range users {
		result[i] = v
	}
	return result, nil
}

func (s *authServiceImpl) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *authServiceImpl) GetUserType(ctx context.Context, userID string) (string, error) {
	return s.userRepo.GetUserType(ctx, userID)
}

func (s *authServiceImpl) ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !helpers.VerifyPassword(oldPassword, *user.Password) {
		return errors.New("old password is incorrect")
	}

	hashedNewPassword, err := helpers.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, userID, hashedNewPassword)
}

func (s *authServiceImpl) AdminChangePassword(ctx context.Context, adminID, targetUserID, newPassword string) error {
	adminType, err := s.userRepo.GetUserType(ctx, adminID)
	if err != nil {
		return errors.New("admin not found")
	}

	if adminType != "SELLER" {
		return errors.New("only seller can change password")
	}

	targetUser, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return errors.New("target user not found")
	}

	if *targetUser.User_type != "USER" {
		return errors.New("target user is not a seller")
	}

	hashedNewPassword, err := helpers.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, targetUserID, hashedNewPassword)
}

func (s *authServiceImpl) Logout(ctx context.Context, userID string, deviceID string) error {
	return helpers.InvalidateRefreshToken(userID, deviceID)
}

func (s *authServiceImpl) LogoutAll(ctx context.Context, userID string) error {
	return helpers.InvalidateAllUserRefreshToken(userID)
}
