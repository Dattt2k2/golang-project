package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"errors"

	database "github.com/Dattt2k2/golang-project/database/databaseConnection.gp"
	// "github.com/Dattt2k2/golang-project/helpers"
	helper "github.com/Dattt2k2/golang-project/auth-service/helpers"
	"github.com/Dattt2k2/golang-project/auth-service/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()



func HashPass(password string) string {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    if err != nil {
        log.Panic(err)
    }
    return string(bytes)
}

func VerifyPass(userPassword string, providedPassword string) (bool, string){
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil{
		msg = fmt.Sprintf("email or password is incorect")
		check = false
	}
	return check, msg
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		// Bind JSON vào user struct
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if user.Email == nil || user.Password == nil || user.First_name == nil || user.Last_name == nil || user.User_type == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email, password, first name, last name and user type are required"})
			return
		}

		// Validate struct
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": validationErr.Error()})
			return
		}

		emailExists ,err := helper.CheckEmailExists(*user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while checking email"})
			return
		}
		if emailExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already taken"})
			return
		}

		// Hash mật khẩu
		password := HashPass(*user.Password)
		user.Password = &password

		// Cập nhật thời gian tạo và cập nhật
		user.Created_at = time.Now()
		user.Updated_at = time.Now()
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		// Tạo token và refresh token
		token, refreshToken, err := helper.GenerateAllToken(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while generating token"})
			return
		}


		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User could not be created"})
			return
		}

		// Trả về token cho người dùng sau khi đăng ký thành công
		c.JSON(http.StatusOK, gin.H{
			"message":       "User created successfully",
			"user":          resultInsertionNumber,
			"access_token":  token,
			"refresh_token": refreshToken,
		})
	}
}

// func SignUp() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
// 		defer cancel()

// 		var user models.User

// 		// Bind JSON vào user struct
// 		if err := c.BindJSON(&user); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		// Validate struct
// 		validationErr := validate.Struct(user)
// 		if validationErr != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"err": validationErr.Error()})
// 			return
// 		}

// 		// Kiểm tra email có tồn tại không
// 		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while checking email"})
// 			return
// 		}
// 		if count > 0 {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already taken"})
// 			return
// 		}

// 		// Kiểm tra số điện thoại có tồn tại không
// 		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while checking phone number"})
// 			return
// 		}
// 		if count > 0 {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is already taken"})
// 			return
// 		}

// 		// Hash mật khẩu
// 		password := HashPass(*user.Password)
// 		user.Password = &password

// 		// Cập nhật thời gian tạo và cập nhật
// 		user.Created_at = time.Now()
// 		user.Updated_at = time.Now()
// 		user.ID = primitive.NewObjectID()
// 		user.User_id = user.ID.Hex()

// 		// Tạo token và refresh token
// 		token, refreshToken, err := helper.GenerateAllToken(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while generating token"})
// 			return
// 		}


// 		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
// 		if insertErr != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "User could not be created"})
// 			return
// 		}

// 		// Trả về token cho người dùng sau khi đăng ký thành công
// 		c.JSON(http.StatusOK, gin.H{
// 			"message":       "User created successfully",
// 			"user":          resultInsertionNumber,
// 			"access_token":  token,
// 			"refresh_token": refreshToken,
// 		})
// 	}
// }


func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User

		// Liên kết dữ liệu JSON vào biến user
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Tìm kiếm người dùng trong cơ sở dữ liệu
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Email or password is incorrect"})
			return
		}

		// Kiểm tra tính hợp lệ của mật khẩu
		passwordValid, msg := VerifyPass(*user.Password, *foundUser.Password) // Đảm bảo so sánh đúng mật khẩu
		if !passwordValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg}) // Trả về lỗi Unauthorized nếu mật khẩu sai
			return
		}

		// Nếu không tìm thấy người dùng
		if foundUser.Email == nil { // Kiểm tra Email, thay vì checking nil
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		// Tạo Access token và Refresh token
		token, refreshToken, err := helper.GenerateAllToken(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens"})
			return
		}

		// Tạo đối tượng LoginResponse với thông tin người dùng và token
		loginResponse := models.LoginResponse{
			Email:        *foundUser.Email,
			First_name:    *foundUser.First_name,
			Last_name:     *foundUser.Last_name,
			User_type:     *foundUser.User_type,
			User_id:       foundUser.User_id,
			Token:        token,
			RefreshToken: refreshToken,
		}

		// Trả về thông tin người dùng cùng token
		c.JSON(http.StatusOK, loginResponse)
	}
}


func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		// Các stages của pipeline
		matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{"_id", "null"},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}},
		}}}
		projectStage := bson.D{{Key: "$project", Value: bson.D{
			{"_id", 0},
			{"total_count", 1},
			{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
		}}}

		// Thực hiện truy vấn aggregate
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred"})
			return
		}

		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allUsers)
	}
}


func GetUser() gin.HandlerFunc{
	return func(c *gin.Context){
		userId := c.Param("user_id")

		if err:= helper.MatchUserTypeToUid(c, userId); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id":userId}).Decode(&user)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

// func GetUserType() gin.HandlerFunc{
// 	return func (c *gin.Context){
// 		userID := c.GetString("uid")
// 		if userID == ""{
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "User Id not found"})
// 		}

// 		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
// 		defer cancel()

// 		var user models.User

// 		err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
// 		if err != nil{
// 			if err == mongo.ErrNoDocuments{
// 				c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
// 			}
// 		}
// 		c.JSON(http.StatusOK, user.User_type)
// 	}
// }

func GetUserType(c *gin.Context) (string, error) {
    userID := c.GetString("uid")
    if userID == "" {
        return "", errors.New("Failed to get uid")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
    defer cancel()

    var result struct {
        UserType string `bson:"user_type"`
    }

    err := userCollection.FindOne(ctx, bson.M{"user_id": userID}, options.FindOne().SetProjection(bson.M{"user_type": 1, "_id": 0})).Decode(&result)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return "", errors.New("User not found")
        }
        return "", err
    }

    return result.UserType, nil
}


func Logout() gin.HandlerFunc {
	return func(c *gin.Context){
		userId := c.GetString("uid")
		deviceId := c.GetHeader("device_id")

		if userId == "" || deviceId == ""{
			c.JSON(http.StatusBadRequest, gin.H{"error": "User Id or Device ID not found"})
			return
		}

		err := helper.InvalidateRefreshToken(userId, deviceId)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when logout"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logout successfully"})
	}
}

func LogoutAll() gin.HandlerFunc{
	return func (c *gin.Context){
		userId := c.GetString("uid")

		if userId == ""{
			c.JSON(http.StatusBadRequest, gin.H{"error": "User Id not found"})
			return
		}

		err := helper.InvalidateAllUserRefreshToken(userId)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when logout"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logout successfully"})
	}
}

func GetDevices() gin.HandlerFunc{
	return func (c *gin.Context){
		userId := c.GetString("uid")

		if userId == ""{
			c.JSON(http.StatusBadRequest, gin.H{"error": "User Id not found"})
			return
		}

		devices, err := helper.GetUserDevices(userId)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when get Devices"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"devices": devices})
	}
}