package helpers

import (
	"auth-service/database"
	"auth-service/logger"
	"context"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Lưu refresh token đã hết hạn vào MongoDB
func SaveExpiredRefreshToken(userID, refreshToken string, expiresAt time.Time) error {
	collection := database.OpenCollection(database.Client, "expired_refresh_tokens")
	filter := bson.M{"user_id": userID, "refresh_token": refreshToken}
	update := bson.M{
		"$set": bson.M{
			"user_id": userID,
			"refresh_token": refreshToken,
			"expires_at": expiresAt,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(context.Background(), filter, update, opts)
	return err
}

// Kiểm tra refresh token đã hết hạn có tồn tại không
func IsExpiredRefreshToken(userID, refreshToken string) bool {
	collection := database.OpenCollection(database.Client, "expired_refresh_tokens")
	filter := bson.M{"user_id": userID, "refresh_token": refreshToken}
	count, err := collection.CountDocuments(context.Background(), filter)
	return err == nil && count > 0
}

type SignedDetails struct{
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Uid          string `json:"uid"`
	UserType     string `json:"user_type"`
	jwt.RegisteredClaims
}

// var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string

// Hàm init để load SECRET_KEY từ .env
func InitDotEnv() {
	// Tải SECRET_KEY từ file .env
	err := godotenv.Load("./auth-service/.env")
	if err != nil {
		logger.Err("Error loading .env file", err)
	}

	// Lấy giá trị SECRET_KEY từ biến môi trường
	SECRET_KEY = os.Getenv("SECRET_KEY")
	if SECRET_KEY == "" {
		logger.Err("SECRET_KEY is not set in .env file", nil)
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


func GenerateToken(email, firstname, lastname, userType, uid string, duration time.Duration)(string, error){
	claims := &SignedDetails{
		Email: email,
		FirstName: firstname,
		LastName: lastname,
		Uid: uid,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(duration)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil{
		return "", err
	}
	
	return token, nil
} 

func GenerateAllToken(email, firstname, lastname, userType, uid string) (string, string,  error){
	accessToken, err := GenerateToken(email, firstname, lastname, userType, uid, time.Hour * 24)
	if err != nil{
		return "", "", err 
	}

	refreshToken, err := GenerateToken(email, firstname, lastname, userType, uid, time.Hour * 168)
	if err != nil{
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateToken kiểm tra token có hợp lệ không
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {

	if strings.Contains(signedToken, "%"){
		decodedToken, err := url.QueryUnescape(signedToken)
		if err != nil{
			logger.Err("Error unescaping token", err)
		}else{
			signedToken = decodedToken
			logger.Debug("Decoded token: ", logger.Str("token", signedToken))	
		}
	}

	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "the token is invalid"
		return
	}

	   if claims.ExpiresAt.Before(time.Now()) {
			   msg = "token is expired"
			   // Lưu refresh token đã hết hạn vào database để kiểm tra về sau
			   _ = SaveExpiredRefreshToken(claims.Uid, signedToken, claims.ExpiresAt.Time)
			   return
	   }

	   return claims, msg
}

func RefreshToken(refreshToken string) (newAccessToken string, msg string) {
	claims, msg := ValidateToken(refreshToken)
	if msg != "" {
		return "", msg
	}

	// Nếu refresh token hợp lệ, tạo lại access token mới
	newAccessToken, _, err := GenerateAllToken(claims.Email, claims.FirstName, claims.LastName, claims.UserType, claims.Uid)
	if err != nil {
		logger.Err("Error generating new access token", err)
		return "", "Error generating new access token"
	}


	return newAccessToken, ""
}

