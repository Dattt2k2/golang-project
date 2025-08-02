package helpers

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"api-gateway/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

type SignedDetails struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Uid       string `json:"uid"`
	UserType  string `json:"user_type"`
	jwt.RegisteredClaims
}

// var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string

// Hàm init để load SECRET_KEY từ .env
func InitDotEnv() {
	// Tải SECRET_KEY từ file .env
	err := godotenv.Load("./api-gateway/.env")
	if err != nil {
		logger.Err("Error loading .env file", err)
	}

	// Lấy giá trị SECRET_KEY từ biến môi trường JWT_SECRET
	SECRET_KEY = os.Getenv("JWT_SECRET")
	if SECRET_KEY == "" {
		logger.Err("SECRET_KEY is not set in .env file", nil)
	} else {
		logger.Debug("SECRET_KEY loaded successfully", logger.Str("SECRET_KEY", SECRET_KEY))
	}
}

// GenerateAllToken tạo access token và refresh token
// func GenerateAllToken(email, firstname, lastname, userType, uid string) (signedToken string, signedRefreshToken string, err error) {
// 	claims := &SignedDetails{
// 		Email:     email,
// 		FirstName: firstname,
// 		LastName:  lastname,
// 		Uid:       uid,
// 		UserType:  userType,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * 24)), // Hết hạn sau 24 giờ
// 		},
// 	}

// 	refreshClaims := &SignedDetails{
// 		Email:     email,
// 		FirstName: firstname,
// 		LastName:  lastname,
// 		Uid:       uid,
// 		UserType:  userType,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * 168)), // Hết hạn sau 7 ngày (168 giờ)
// 		},
// 	}

// 	// Tạo JWT token
// 	signedToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRECT_KEY))
// 	if err != nil {
// 		log.Panic(err)
// 		return
// 	}

// 	signedRefreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRECT_KEY))
// 	if err != nil {
// 		log.Panic(err)
// 		return
// 	}

// 	return signedToken, signedRefreshToken, err
// }

func GenerateToken(email, firstname, lastname, userType, uid string, duration time.Duration) (string, error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstname,
		LastName:  lastname,
		Uid:       uid,
		UserType:  userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(duration)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	return token, nil
}

func GenerateAllToken(email, firstname, lastname, userType, uid string) (string, string, error) {
	accessToken, err := GenerateToken(email, firstname, lastname, userType, uid, time.Hour*24)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := GenerateToken(email, firstname, lastname, userType, uid, time.Hour*168)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateToken kiểm tra token có hợp lệ không
// func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
// 	// Check if token is empty
// 	if signedToken == "" {
// 		logger.Debug("Token is empty")
// 		return nil, "token is empty"
// 	}

// 	// URL-decode if needed
// 	if strings.Contains(signedToken, "%") {
// 		decodedToken, err := url.QueryUnescape(signedToken)
// 		if err != nil {
// 			logger.DebugE("Failed to decode token: %v", err)
// 		} else {
// 			signedToken = decodedToken
// 			logger.DebugE("Token decoded successfully", nil)
// 		}
// 	}

// 	// Check if token has correct format
// 	parts := strings.Split(signedToken, ".")
// 	if len(parts) != 3 {

// 		logger.DebugE("Malformed token", nil)
// 		return nil, "token contains an invalid number of segments"
// 	}

// 	// Parse the token
// 	token, err := jwt.ParseWithClaims(
// 		signedToken,
// 		&SignedDetails{},
// 		func(token *jwt.Token) (interface{}, error) {
// 			return []byte(SECRET_KEY), nil
// 		},
// 	)

// 	if err != nil {
// 		logger.DebugE("Failed to parse token", err)
// 		msg = err.Error()
// 		return
// 	}

// 	claims, ok := token.Claims.(*SignedDetails)
// 	if !ok {
// 		logger.Debug("Invalid claims type")
// 		msg = "the token is invalid"
// 		return
// 	}

// 	if claims.ExpiresAt.Before(time.Now().UTC()) {
// 		logger.Debug("Token is expired")
// 		msg = "token is expired"
// 		// RefreshToken(signedToken)
// 		return
// 	}

// 	logger.DebugE("Token valid for user", nil, logger.Str("email", claims.Email), logger.Str("user_type", claims.UserType))
// 	return claims, ""
// }

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	fmt.Println("[DEBUG] Token to validate:", signedToken)
	fmt.Println("[DEBUG] SECRET_KEY for validation:", SECRET_KEY)
	logger.Debug("Starting token validation", logger.Str("token", signedToken))
	logger.Debug("Using SECRET_KEY for validation", logger.Str("SECRET_KEY", SECRET_KEY))

	// Check if token is empty
	if signedToken == "" {
		logger.Debug("Token is empty")
		return nil, "token is empty"
	}

	// URL-decode if needed
	if strings.Contains(signedToken, "%") {
		decodedToken, err := url.QueryUnescape(signedToken)
		if err != nil {
			logger.DebugE("Failed to decode token", err)
		} else {
			signedToken = decodedToken
			logger.Debug("Token decoded successfully", logger.Str("decodedToken", signedToken))
		}
	}

	// Check if token has correct format
	parts := strings.Split(signedToken, ".")
	if len(parts) != 3 {
		logger.Debug("Malformed token", logger.Str("token", signedToken))
		return nil, "token contains an invalid number of segments"
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			logger.Debug("Parsing token with claims", logger.Str("method", token.Method.Alg()))

			// Check signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				logger.Debug("Unexpected signing method", logger.Str("method", token.Method.Alg()))
				return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
			}

			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		logger.DebugE("Failed to parse token", err)
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		logger.Debug("Invalid claims type")
		msg = "the token is invalid"
		return
	}

	if !token.Valid {
		logger.Debug("Token is not valid")
		msg = "the token is invalid"
		return
	}

	if claims.ExpiresAt.Before(time.Now().UTC()) {
		logger.Debug("Token is expired", logger.Str("expiresAt", claims.ExpiresAt.String()))
		msg = "token is expired"
		return
	}

	logger.Debug("Token validated successfully", logger.Str("email", claims.Email), logger.Str("user_type", claims.UserType))
	return claims, ""
}

func RefreshToken(refreshToken string) (newAccessToken string, msg string) {
	claims, msg := ValidateToken(refreshToken)
	if msg != "" {
		return "", msg
	}

	// Nếu refresh token hợp lệ, tạo lại access token mới
	newAccessToken, _, err := GenerateAllToken(claims.Email, claims.FirstName, claims.LastName, claims.UserType, claims.Uid)
	if err != nil {
		return "", "Error generating new access token"
	}

	return newAccessToken, ""
}

func ExtractBearerToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization header is missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("Authorization header format must be Bearer {token}")
	}
	return parts[1], nil
}

func ValidateBearerToken(c *gin.Context) (*SignedDetails, error) {
	token, err := ExtractBearerToken(c)
	if err != nil {
		return nil, err
	}

	claims, msg := ValidateToken(token)
	if msg != "" {
		return nil, errors.New(msg)
	}
	return claims, nil
}

// Helper function to get the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
