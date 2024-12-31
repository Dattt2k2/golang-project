// package helpers

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"os"
// 	"time"

// 	database "github.com/Dattt2k2/golang-project/database/databaseConnection.gp"
// 	"github.com/golang-jwt/jwt/v4"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// type SignedDetails struct{
// 	Email 			string
// 	First_name 		string
// 	Last_name		string
// 	Uid				string
// 	User_type 		string
// 	jwt.RegisteredClaims
// }

// var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

// var SECRECT_KEY string = os.Getenv("SECRECT_KEY")

// func GenerateAllToken(email string, firstname string, lastname string, userType string, uid string) (signedToken string, signedRefreshToken string, err error){
// 	claims := &SignedDetails{
// 		Email : email,
// 		First_name: firstname,
// 		Last_name: lastname,
// 		Uid: uid,
// 		User_type: userType,
// 		RegisteredClaims: jwt.RegisteredClaims{
//             ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * time.Duration(24))),
//         },
// 	}
// 	refreshClaims := &SignedDetails{
// 		RegisteredClaims: jwt.RegisteredClaims{
//             ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * time.Duration(168))),
//         },
// 	}

// 	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRECT_KEY))
// 	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRECT_KEY))

// 	if err != nil{
// 		log.Panic(err)
// 		return
// 	}
// 	return token, refreshToken, err
// }

// func ValidateToken(signedToken string) (claims *SignedDetails, msg string){
// 	token, err :=  jwt.ParseWithClaims(
// 		signedToken,
// 		&SignedDetails{},
// 		func(token *jwt.Token)(interface{}, error){
// 			return []byte(SECRECT_KEY), nil
// 		},
// 	)
// 	if err != nil{
// 		msg = err.Error()
// 		return
// 	}
// 	claims, ok := token.Claims.(*SignedDetails)
// 	if !ok{
// 		msg = fmt.Sprintf("the token is invalid")
// 		msg = err.Error()
// 		return
// 	}
// 	if claims.ExpiresAt.Before(time.Now()){
// 		msg = fmt.Sprintf("token is expired")
// 		msg = err.Error()
// 		return
// 	}
// 	return claims, msg
// }

// func UpdateAllToken(signedToken string, signedRefreshToken string, userId string){
// 	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
// 	var updateObj primitive.D

// 	updateObj = append(updateObj, bson.E{"token", signedToken})
// 	updateObj = append (updateObj, bson.E{"refresh_token", signedRefreshToken})

// 	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
// 	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})
// 	upsert := true
// 	filter := bson.M{"user_id": userId}
// 	opt := options.UpdateOptions{
// 		Upsert : &upsert,
// 	}
// 	_, err := userCollection.UpdateOne(
// 		ctx,
// 		filter,
// 		bson.D{
// 			{"$set", updateObj},
// 		},
// 		&opt,
// 	) 

// 	defer cancel()	

// 	if err != nil{
// 		log.Panic(err)
// 		return
// 	}
// 	return
// }

package helpers

import (
	// "context"
	// "fmt"
	"log"
	"os"
	"time"

	// database "github.com/Dattt2k2/golang-project/database/databaseConnection.gp"
	"github.com/golang-jwt/jwt/v4"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
	"github.com/joho/godotenv"
)

type SignedDetails struct{
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Uid          string `json:"uid"`
	UserType     string `json:"user_type"`
	jwt.RegisteredClaims
}

// var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRECT_KEY string

// Hàm init để load SECRET_KEY từ .env
func InitDotEnv() {
	// Tải SECRET_KEY từ file .env
	err := godotenv.Load("./auth-service/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Lấy giá trị SECRET_KEY từ biến môi trường
	SECRECT_KEY = os.Getenv("SECRET_KEY")
	if SECRECT_KEY == "" {
		log.Fatal("SECRET_KEY not found in .env")
	}
}

// GenerateAllToken tạo access token và refresh token
func GenerateAllToken(email, firstname, lastname, userType, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstname,
		LastName:  lastname,
		Uid:       uid,
		UserType:  userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * 24)), // Hết hạn sau 24 giờ
		},
	}

	refreshClaims := &SignedDetails{
		Email:     email,
		FirstName: firstname,
		LastName:  lastname,
		Uid:       uid,
		UserType:  userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * 168)), // Hết hạn sau 7 ngày (168 giờ)
		},
	}

	// Tạo JWT token
	signedToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRECT_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	signedRefreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRECT_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	return signedToken, signedRefreshToken, err
}

// ValidateToken kiểm tra token có hợp lệ không
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRECT_KEY), nil
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
		return "", "Error generating new access token"
	}

	return newAccessToken, ""
}