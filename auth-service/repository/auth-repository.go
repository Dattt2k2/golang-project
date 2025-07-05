package repository

import (
	"context"
	"errors"
	"time"

	"auth-service/database"
	"auth-service/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	Create(ctx context.Context, user *models.User) (*mongo.InsertOneResult, error)
	GetAllUsers(ctx context.Context, startIndex int, recordPerPage int) ([]bson.M, error)
	UpdatePassword(ctx context.Context, userID string, hashedPass string) error 
	GetUserType(ctx context.Context, userID string) (string, error)
}

type userRepositoryImpl struct {
	collection *mongo.Collection
}

func NewUserRepository() UserRepository {
	return &userRepositoryImpl{
		collection: database.OpenCollection(database.Client, "user"),
	}
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User 
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err 
	}
	

	return &user, nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User 
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err 
	}
	return &user, nil
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *models.User) (*mongo.InsertOneResult, error) {
	result, err := r.collection.InsertOne(ctx, user)
	return result, err
}

func (r *userRepositoryImpl) GetAllUsers(ctx context.Context, startIndex int, recordPerPage int) ([]bson.M, error) {

	matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
	groupStage := bson.D{{Key: "$group", Value: bson.D{
		{Key: "_id", Value: "null"},
		{Key: "totalCount", Value: bson.D{{Key: "$sum", Value: 1}}},
		{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
	}}}
	projectStage := bson.D{{Key: "$project", Value: bson.D{
        {Key: "_id", Value: 0},
        {Key: "total_count", Value: 1},
        {Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
    }}}

	result, err := r.collection.Aggregate(ctx, mongo.Pipeline{
		matchStage, groupStage, projectStage,
	})

	if err != nil {
		return nil, err
	}

	var allUsers []bson.M
	if err = result.All(ctx, &allUsers); err != nil {
		return nil, err 
	}
	return allUsers, nil
}

func (r *userRepositoryImpl) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"password": hashedPassword,
			"updated_at": time.Now(),
		},
	}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err 
}

func (r *userRepositoryImpl) GetUserType(ctx context.Context, userID string) (string, error) {
	var result struct {
		UserType string `bson:"user_type"`
	}

	err := r.collection.FindOne(
		ctx,
		bson.M{"_id": userID},
		options.FindOne().SetProjection(bson.M{"user_type": 1, "_id": 0}),
	).Decode(&result)

	if err != nil {
		return "", err 
	}

	return result.UserType, nil
}