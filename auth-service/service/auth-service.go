package service

import (
	"context"
	"errors"
	"time"

	"auth-service/helpers"
	"auth-service/kafka"
	"auth-service/logger"
	"auth-service/models"
	"auth-service/repository"
	"auth-service/websocket"
)

type AuthService interface {
	Register(ctx context.Context, email, password, phone, userType, firstName string) (*models.SignUpResponse, error)
	Login(ctx context.Context, credential *models.LoginCredentials) (*models.LoginResponse, error)
	GetAllUsers(ctx context.Context, page int, recordPage int) ([]interface{}, error)
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserType(ctx context.Context, userID string) (string, error)
	ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error
	AdminChangePassword(ctx context.Context, adminID, targetUserID, newPassword string) error
	Logout(ctx context.Context, userID string, deviceId string) error
	LogoutAll(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	VerifyOTP(ctx context.Context, email, otpCode string) error
	SendOTP(ctx context.Context, email string) (string, error)
	ResendOTP(ctx context.Context, email string) (string, error)
	UpdateUserRole(ctx context.Context, userID, newRole string) error
	UpdateUserDisabled(ctx context.Context, userID string, isDisabled bool) error
	DeleteUser(ctx context.Context, userID string) error
	ForgotPassword(ctx context.Context, email string) error
}

type authServiceImpl struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authServiceImpl{
		userRepo: userRepo,
	}
}

func (s *authServiceImpl) Register(ctx context.Context, email, password, phone, userType, firstName string) (*models.SignUpResponse, error) {
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

	defaultFirstName := "User"
	defaultLastName := ""

	user := &models.User{
		Email:     &email,
		Password:  &hashedPassword,
		FirstName: &firstName,
		LastName:  &defaultLastName,
		UserType:  userType,
		Phone:     &phone,
		IsVerify:  true,
	}
	result, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	userID := result.ID.String()

	token, refreshToken, err := helpers.GenerateAllToken(email, defaultFirstName, defaultLastName, userType, userID)
	if err != nil {
		return nil, err
	}

	err = helpers.AddUserToBloomFilter(*user.Email, *user.Phone)
	if err != nil {
		return nil, err
	}

	// Generate OTP
	// otp, err := helpers.GenerateAndStoreOTP(*user.Email, 5*time.Minute)
	// if err != nil {
	// 	return nil, errors.New("failed to generate OTP")
	// }

	// msg := kafka.EmailMessage{
	// 	To:       *user.Email,
	// 	Subject:  "Welcome to Our Service",
	// 	Template: "./template/otp_send.html",
	// 	Data: map[string]interface{}{
	// 		"FirstName": *user.FirstName,
	// 		"LastName":  *user.LastName,
	// 		"Email":     *user.Email,
	// 		"OTP":       otp,
	// 	},
	// }

	// err = kafka.SendEmailMessage(kafka.NewKafkaWriter("kafka:9092", "email_topic"), msg)
	// if err != nil {
	// 	return nil, errors.New("failed to send otp")
	// }
	// publish user.created event for downstream services
	userEvent := map[string]interface{}{
		"id":         result.ID.String(),
		"email":      *user.Email,
		"first_name": *user.FirstName,
		"last_name":  *user.LastName,
		"phone":      *user.Phone,
		"user_type":  user.UserType,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	err = kafka.SendJSONMessage(kafka.NewKafkaWriter("kafka:9092", "user.created"), userEvent)
	if err != nil {
		return nil, err
	}
	logger.Info("Published user.created event to Kafka")
	return &models.SignUpResponse{
		Message:      "User registered successfully",
		User:         result,
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authServiceImpl) Login(ctx context.Context, credential *models.LoginCredentials) (*models.LoginResponse, error) {
	foundUser, err := s.userRepo.FindByEmail(ctx, *credential.Email)
	if err != nil || foundUser == nil {
		return nil, errors.New("invalid email or password")
	}

	if !helpers.CheckIsVerify(foundUser) {
		return nil, errors.New("user is not verified")
	}

	if !helpers.VerifyPassword(*credential.Password, *foundUser.Password) {
		return nil, errors.New("email or password is incorrect")
	}

	token, refreshToken, err := helpers.GenerateAllToken(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, foundUser.UserType, foundUser.ID.String())
	if err != nil {
		return nil, errors.New("error generating token")
	}

	loginResponse := &models.LoginResponse{
		Email:        *foundUser.Email,
		First_name:   *foundUser.FirstName,
		Last_name:    *foundUser.LastName,
		User_type:    foundUser.UserType,
		User_id:      foundUser.ID.String(),
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

func (s *authServiceImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.FindByEmailAny(ctx, email)
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

	if targetUser.UserType != "USER" {
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

func (s *authServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (string, error) {

	claims, msg := helpers.ValidateToken(refreshToken)
	if msg != "" {
		if claims != nil && helpers.IsExpiredRefreshToken(claims.Uid, refreshToken) {
			return "", errors.New("refresh token is expired and cannot be used to generate a new access token")
		}
		return "", errors.New(msg)
	}

	accessToken, msg := helpers.RefreshToken(refreshToken)
	if msg != "" {
		return "", errors.New(msg)
	}
	return accessToken, nil
}

func (s *authServiceImpl) VerifyOTP(ctx context.Context, email, otpCode string) error {
	otp, err := helpers.GetOTP(email)
	if err != nil {
		return errors.New("failed to retrieve OTP code")
	}

	if otp != otpCode {
		return errors.New("invalid OTP code")
	}

	err = s.userRepo.UpdateVerificationStatus(ctx, email, true)
	if err != nil {
		return errors.New("failed to update user verification status")
	}
	return nil
}

func (s *authServiceImpl) SendOTP(ctx context.Context, email string) (string, error) {
	otp, err := helpers.GenerateAndStoreOTP(email, 5*time.Minute)
	if err != nil {
		return "", errors.New("failed to generate OTP")
	}
	msg := kafka.EmailMessage{
		To:       email,
		Subject:  "Your OTP Code",
		Template: "./template/otp_send.html",
		Data: map[string]interface{}{
			"Email": email,
			"OTP":   otp,
		},
	}
	err = kafka.SendEmailMessage(kafka.NewKafkaWriter("kafka:9092", "email_topic"), msg)
	if err != nil {
		return "", errors.New("failed to send OTP email")
	}

	return otp, nil
}

func (s *authServiceImpl) ResendOTP(ctx context.Context, email string) (string, error) {
	otp, err := helpers.ResendOTP(email)
	if err != nil {
		return "", errors.Join()
	}

	msg := kafka.EmailMessage{
		To:       email,
		Subject:  "Resend OTP",
		Template: "./template/otp_send.html",
		Data: map[string]interface{}{
			"Email": email,
			"OTP":   otp,
		},
	}

	err = kafka.SendEmailMessage(kafka.NewKafkaWriter("kafka:9092", "email_topic"), msg)
	if err != nil {
		return "", errors.New("failed to send OTP email")
	}

	return otp, nil
}

func (s *authServiceImpl) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("user with the provided email does not exist")
	}

	otp, err := helpers.GenerateAndStoreOTP(email, 10*time.Minute)
	if err != nil {
		return errors.New("failed to generate OTP")
	}

	msg := kafka.EmailMessage{
		To:       email,
		Subject:  "Password Reset OTP",
		Template: "./template/otp_send.html",
		Data: map[string]interface{}{
			"FirstName": *user.FirstName,
			"LastName":  *user.LastName,
			"OTP":       otp,
		},
	}

	err = kafka.SendEmailMessage(kafka.NewKafkaWriter("kafka:9092", "email_topic"), msg)
	if err != nil {
		return errors.New("failed to send OTP email")
	}

	return nil
}

func (s *authServiceImpl) UpdateUserRole(ctx context.Context, userID, newRole string) error {
	err := s.userRepo.UpdateRole(ctx, userID, newRole)
	if err != nil {
		return err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	accessToken, err := helpers.GenerateToken(*user.Email, *user.FirstName, *user.LastName, newRole, userID, time.Hour*24)
	if err != nil {
		return err
	}

	refreshToken, err := helpers.GenerateToken(*user.Email, *user.FirstName, *user.LastName, newRole, userID, time.Hour*168)
	if err != nil {
		return err
	}

	err = websocket.NotifyRoleChange(userID, newRole, accessToken, refreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (s *authServiceImpl) UpdateUserDisabled(ctx context.Context, userID string, isDisabled bool) error {
	return s.userRepo.UpdateDisabled(ctx, userID, isDisabled)
}

func (s *authServiceImpl) DeleteUser(ctx context.Context, userID string) error {
	return s.userRepo.DeleteUser(ctx, userID)
}
