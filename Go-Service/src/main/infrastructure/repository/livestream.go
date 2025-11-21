package repository

import (
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/livestream"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoLivestreamRepository struct {
	collection *mongo.Collection
}

func NewMongoLivestreamRepository(db *mongo.Database) repository.LivestreamRepository {
	return &MongoLivestreamRepository{
		collection: db.Collection("livestreams"),
	}
}

func (r *MongoLivestreamRepository) GetByID(id string) (*livestream.Livestream, error) {
	var ls livestream.Livestream
	err := r.collection.FindOne(context.Background(), bson.M{"uuid": id}).Decode(&ls)
	return &ls, err
}

func (r *MongoLivestreamRepository) GetByOwnerID(ownerID string) (*livestream.Livestream, error) {
	var ls livestream.Livestream
	err := r.collection.FindOne(context.Background(), bson.M{"owneruserid": ownerID}).Decode(&ls)
	return &ls, err
}

func (r *MongoLivestreamRepository) GetOne() (*livestream.Livestream, error) {
	var ls livestream.Livestream
	err := r.collection.FindOne(context.Background(), bson.M{}).Decode(&ls)
	if err == mongo.ErrNoDocuments {
		return nil, errors.ErrNotFound
	}
	return &ls, err
}

func (r *MongoLivestreamRepository) Create(ls *livestream.Livestream) error {
	_, err := r.collection.InsertOne(context.Background(), ls)
	return err
}

func (r *MongoLivestreamRepository) Update(ls *livestream.Livestream) error {
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"uuid": ls.UUID}, bson.M{"$set": ls})
	return err
}

func (r *MongoLivestreamRepository) Delete(id string) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"uuid": id})
	return err
}
func (r *MongoLivestreamRepository) MuteUser(identityProvider string, livestreamUUID string, userID string) error {
	// Add the userID to the MuteList array
	update := bson.M{"$addToSet": bson.M{"mutelist": identityProvider + "-" + userID}}
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"uuid": livestreamUUID}, update)
	return err
}
