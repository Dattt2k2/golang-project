package helpers

import (
	"context"
	"time"

	// database "auth-service/database"
	"auth-service/logger"
	"auth-service/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var userDB *gorm.DB
var userBloom *BloomFilter

func SetDB(db *gorm.DB) {
	userDB = db
}

func SetUserBloomFilter(bf *BloomFilter){
	userBloom = bf
}

func CheckUsernameExists(username string) (bool, error){
	if userBloom != nil{
		exists, err := userBloom.Contains(username)
		if err != nil {
			logger.Err("Error checking username in BloomFilter", err)
			return false, err 
		}
		if !exists{
			return false, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var count int64
	if err := userDB.WithContext(ctx).Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		logger.Err("Error checking username in Postgres", err)
		return false, err
	}
	return count > 0, nil
}


func CheckEmailExists(email string) (bool, error){
	if userBloom != nil {
        exists, err := userBloom.Contains(email)
        if err != nil {
            logger.Err("Error checking email in BloomFilter", err)
        } else if !exists {
            return false, nil
        }
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var count int64
    if err := userDB.WithContext(ctx).Model(&models.User{}).
        Where("email = ?", email).
        Count(&count).Error; err != nil {
        logger.Err("Error checking email in Postgres", err)
        return false, err
    }

    return count > 0, nil
}

func AddUserToBloomFilter(email, phone string) error {
    if userBloom == nil {
        return nil  
    }
    
    if email != "" {
        userBloom.Add(email)
    }
    
    if phone != "" {
        userBloom.Add(phone)
    }
	return nil
}

// func CheckUserType(c *gin.Context, role string) (err error){
// 	userType := c.GetString("user_type")
// 	err = nil
// 	if userType != role{
// 		err = errors.New("Unauthorized to acces the resource")
// 		return err
// 	}
// 	return err
// }



// func MatchUserTypeToUid(c *gin.Context, userId string) (err error){
// 	userType := c.GetString("user_type")
// 	uid := c.GetString("uid")
// 	err = nil

// 	if userType == "USER" && uid != userId{
// 		err = errors.New("Unauthorized to access the resource")
// 		return err
// 	}
// 	err = CheckUserType(c, userType)
// 	return err
// }

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil  {
		return "", err 
	}
	return string(bytes), nil
}

func VerifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func CheckIsVerify(user *models.User) bool {
	if user == nil {
		return false 
	}
	return user.IsVerify
}