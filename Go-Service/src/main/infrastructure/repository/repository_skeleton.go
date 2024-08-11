// Go-Service/src/main/infrastructure/repository/mongo_skeleton_repository.go
package repository

import (
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/domain/entity"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoSkeletonRepository struct {
	collection *mongo.Collection
}

func NewMongoSkeletonRepository(db *mongo.Database) repository.SkeletonRepository {
	return &MongoSkeletonRepository{
		collection: db.Collection("skeletons"),
	}
}

func (r *MongoSkeletonRepository) GetByID(id string) (*entity.Skeleton, error) {
	var skeleton entity.Skeleton
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&skeleton)
	return &skeleton, err
}

func (r *MongoSkeletonRepository) Create(skeleton *entity.Skeleton) error {
	_, err := r.collection.InsertOne(context.Background(), skeleton)
	return err
}
