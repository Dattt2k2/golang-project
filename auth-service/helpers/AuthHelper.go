package helpers

import (
	"context"
	"errors"
	"log"
	"time"

	database "github.com/Dattt2k2/golang-project/auth-service/database"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var userBloom *BloomFilter

func SetUserBloomFilter(bf *BloomFilter){
	userBloom = bf
}

func CheckUsernameExists(username string) (bool, error){
	exists, err := userBloom.Contains(username)
	if err != nil{
		log.Printf("Error checking username: %v", err)
		return false, err
	}

	if !exists{
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	count, err := userCollection.CountDocuments(ctx, bson.M{"username": username})
	if err != nil{
		return false, err
	}

	return count > 0, nil
}

func CheckEmailExists(email string) (bool, error){
	exists , err := userBloom.Contains(email)
	if err != nil{
		log.Printf("Error checking email: %v", err)
		return false, err
	}

	if !exists{
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil{
		return false, err
	}

	return count > 0, nil
}

func AddUserToBloomFilter(email, username, phone string) {
    if userBloom == nil {
        return
    }
    
    if email != "" {
        userBloom.Add(email)
    }
    
    if username != "" {
        userBloom.Add(username)
    }
    
    if phone != "" {
        userBloom.Add(phone)
    }
}

func CheckUserType(c *gin.Context, role string) (err error){
	userType := c.GetString("user_type")
	err = nil
	if userType != role{
		err = errors.New("Unauthorized to acces the resource")
		return err
	}
	return err
}

func MatchUserTypeToUid(c *gin.Context, userId string) (err error){
	userType := c.GetString("user_type")
	uid := c.GetString("uid")
	err = nil

	if userType == "USER" && uid != userId{
		err = errors.New("Unauthorized to access the resource")
		return err
	}
	err = CheckUserType(c, userType)
	return err
}