package repository

import (
	"context"

	"github.com/Dattt2k2/golang-project/product-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type ProductRepository interface {
	Insert(ctx context.Context, product models.Product) error
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	Delete(ctx context.Context, id, userID primitive.ObjectID) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Product, error)
	FindByName(ctx context.Context, name string) ([]models.Product, error)
	FindAll(ctx context.Context, skip, limit int64) ([]models.Product, int64, error)
	UpdateStock(ctx context.Context, id primitive.ObjectID, quantity int) error
}

type productRepositoryImpl struct {
	collection *mongo.Collection
}

func NewProductRepository(collection *mongo.Collection) ProductRepository {
	return &productRepositoryImpl{collection: collection}
}

func (r *productRepositoryImpl) Insert(ctx context.Context, product models.Product) error {
	_, err := r.collection.InsertOne(ctx, product)
	return err
}

func (r *productRepositoryImpl) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	_, err  := r.collection.UpdateOne(ctx, bson.M{"_id": id},bson.M{"$set": update})
	return err 
}

func (r *productRepositoryImpl) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	filter := bson.M{"_id": id, "user_id": userID}
	res, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err 
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil 
}

func (r *productRepositoryImpl) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Product, error) {
	var product models.Product
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err  != nil {
		return nil, err 
	}
	return &product, nil
}

func (r *productRepositoryImpl) FindByName(ctx context.Context, name string) ([]models.Product, error) {
	var products []models.Product
	filter := bson.M{"name":bson.M{"$regex": name, "$options": "i"}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err 
	}

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err 
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *productRepositoryImpl) FindAll(ctx context.Context, skip, limit int64) ([]models.Product, int64, error){
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err 
	}
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts) 
	if err != nil{
		return nil, 0, err 
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil{
		return nil, 0, err 
	}

	return products, total, nil 
}

func (r *productRepositoryImpl) UpdateStock(ctx context.Context, id primitive.ObjectID, quantity int) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"quantity": quantity}})
	return err 
}