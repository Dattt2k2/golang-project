package helpers

import (
	// "context"
	// "fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

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
		log.Fatal("Error loading .env file")
	}

	// Lấy giá trị SECRET_KEY từ biến môi trường
	SECRET_KEY = os.Getenv("SECRET_KEY")
	if SECRET_KEY == "" {
		log.Fatal("SECRET_KEY not found in .env")
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
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	// Check if token is empty
	if signedToken == "" {
		log.Println("[DEBUG] Empty token provided")
		return nil, "token is empty"
	}

	// Add debug logging
	log.Printf("[DEBUG] Validating token (length: %d)", len(signedToken))
	log.Printf("[DEBUG] SECRET_KEY length: %d", len(SECRET_KEY))

	// URL-decode if needed
	if strings.Contains(signedToken, "%") {
		decodedToken, err := url.QueryUnescape(signedToken)
		if err != nil {
			log.Println("[DEBUG] Error unescaping token: ", err)
		} else {
			signedToken = decodedToken
			log.Printf("[DEBUG] Decoded token (length: %d)", len(signedToken))
		}
	}

	// Check if token has correct format
	parts := strings.Split(signedToken, ".")
	if len(parts) != 3 {
		log.Printf("[DEBUG] Malformed token: has %d parts instead of 3", len(parts))
		return nil, "token contains an invalid number of segments"
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		log.Printf("[DEBUG] JWT parse error: %v", err)
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		log.Printf("[DEBUG] Invalid claims type")
		msg = "the token is invalid"
		return
	}

	if claims.ExpiresAt.Before(time.Now()) {
		log.Printf("[DEBUG] Token expired")
		msg = "token is expired"
		return
	}

	log.Printf("[DEBUG] Token valid for user: %s", claims.Email)
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

// Helper function to get the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
