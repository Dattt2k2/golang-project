package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"auth-service/helpers"
	"auth-service/logger"
	"auth-service/models"
	"auth-service/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

type AuthController struct {
	authService service.AuthService
	validate    *validator.Validate
}

func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
		validate:    validator.New(),
	}
}

func (ctrl *AuthController) SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		var user models.User

		if err := c.ShouldBindJSON(&user); err != nil {
			logger.Err("Error binding JSON", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if user.Email == nil || user.Password == nil {
			logger.Err("Email or password is nil", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email, password are required"})
			return
		}

		validationErr := ctrl.validate.Struct(user)
		if validationErr != nil {
			logger.Err("Validation error", validationErr)
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		response, err := ctrl.authService.Register(ctx, *user.Email, *user.Password, *user.User_type)
		if err != nil {
			logger.Err("Error registering user", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = SendOTPHander(*user.Email, "send_otp.html")
		if err != nil {
			logger.Error("Error sending OTP", logger.ErrField(err))
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "User created successfully",
			"user":          response.User,
			"access_token":  response.Token,
			"refresh_token": response.RefreshToken,
		})
	}
}

func (ctrl *AuthController) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		var credential models.LoginCredentials

		if err := c.ShouldBindJSON(&credential); err != nil {
			logger.Err("Error binding JSON", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := ctrl.authService.Login(ctx, &credential)
		if err != nil {
			logger.Err("Error logging in", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		deviceID := c.GetHeader("X-Device-ID")
		platform := c.GetHeader("X-Platform")
		userAgent := c.GetHeader("User-Agent")
		ipAddress := c.ClientIP()

		if deviceID != "" && response.User_id != "" {
			helpers.StoreRefreshToken(
				response.User_id,
				response.RefreshToken,
				deviceID,
				platform,
				userAgent,
				ipAddress,
			)
		}

		c.JSON(http.StatusOK, response)
		logger.Info("User logged in successfully", logger.Str("email", *credential.Email))
	}
}

func CheckSellerRole(c *gin.Context) {
	userRole := c.GetHeader("user_type")
	if userRole != "SELLER" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Your don't have permission"})
		c.Abort()
		return
	}
}

func (ctrl *AuthController) GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		CheckSellerRole(c)
		if c.IsAborted() {
			logger.Err("Unauthorized access", nil)
			return
		}

		recordPage, err := strconv.Atoi(c.Query("limit"))
		if err != nil {
			recordPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			page = 1
		}

		users, err := ctrl.authService.GetAllUsers(ctx, page, recordPage)
		if err != nil {
			logger.Err("Error getting users", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting users"})
			return
		}

		c.JSON(http.StatusOK, users)
		logger.Info("Users retrieved successfully")
	}
}

func (ctrl *AuthController) GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		userId := c.Param("user_id")
		if userId == "" {
			logger.Err("User ID not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			return
		}

		user, err := ctrl.authService.GetUser(ctx, userId)
		if err != nil {
			logger.Err("Error getting user", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting user"})
			return
		}

		c.JSON(http.StatusOK, user)
		logger.Info("User retrieved successfully")
	}
}

func (ctrl *AuthController) ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		var request models.ChangePasswordRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Err("Error binding JSON", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetHeader("user_id")
		if userID == "" {
			logger.Err("User ID not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			return
		}

		err := ctrl.authService.ChangePassword(ctx, userID, request.OldPassword, request.NewPassword)
		if err != nil {
			logger.Err("Error changing password", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error changing password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
		logger.Info("Password changed successfully", logger.Str("user_id", userID))
	}
}

func (ctrl *AuthController) AdminChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		CheckSellerRole(c)
		if c.IsAborted() {
			return
		}

		var request models.AdminChangePassword
		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Err("Error binding JSON", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		adminID := c.GetHeader("user_id")
		if adminID == "" {
			logger.Err("Admin ID not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Admin ID not found"})
			return
		}

		err := ctrl.authService.AdminChangePassword(ctx, adminID, request.UserID, request.NewPassword)
		if err != nil {
			logger.Err("Error changing password", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error changing password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
		logger.Info("Password changed successfully", logger.Str("admin_id", adminID), logger.Str("user_id", request.UserID))
	}
}

func (ctrl *AuthController) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		userID := c.GetHeader("user_id")
		deviceID := c.GetHeader("device_id")

		if userID == "" || deviceID == "" {
			logger.Err("User ID or Device ID not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID or Device ID not found"})
			return
		}

		err := ctrl.authService.Logout(ctx, userID, deviceID)
		if err != nil {
			logger.Err("Error logging out", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error logging out"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
		logger.Info("User logged out successfully", logger.Str("user_id", userID), logger.Str("device_id", deviceID))
	}
}

func (ctrl *AuthController) LogoutAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		userID := c.GetHeader("user_id")
		if userID == "" {
			logger.Err("User ID not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			return
		}

		err := ctrl.authService.LogoutAll(ctx, userID)
		if err != nil {
			logger.Err("Error logging out from all devices", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error logging out from all devices"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logged out from all devices successfully"})

		logger.Info("User logged out from all devices successfully", logger.Str("user_id", userID))
	}
}

func (ctrl *AuthController) GetDevices() gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := c.GetHeader("user_id")
		if userID == "" {
			logger.Err("User ID not found", nil)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			return
		}

		devices, err := helpers.GetUserDevices(userID)
		if err != nil {
			logger.Err("Error getting devices", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting devices"})
			return
		}

		c.JSON(http.StatusOK, devices)
		logger.Info("Devices retrieved successfully", logger.Str("user_id", userID))
	}
}
