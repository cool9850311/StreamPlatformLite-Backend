package repository

import (
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/domain/entity/system"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoSystemSettingRepository struct {
	collection *mongo.Collection
}

func NewMongoSystemSettingRepository(db *mongo.Database) repository.SystemSettingRepository {
	return &MongoSystemSettingRepository{
		collection: db.Collection("system_settings"),
	}
}

func (r *MongoSystemSettingRepository) GetSetting() (*system.Setting, error) {
	var setting system.Setting
	err := r.collection.FindOne(context.Background(), bson.M{}).Decode(&setting)
	return &setting, err
}

func (r *MongoSystemSettingRepository) SetSetting(setting *system.Setting) error {
	// Check if a setting document exists
	count, err := r.collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	// If no document exists, insert a new one
	if count == 0 {
		_, err = r.collection.InsertOne(context.Background(), setting)
		if err != nil {
			return err
		}
		return nil
	}
	_, errr := r.collection.UpdateOne(context.Background(), bson.M{}, bson.M{"$set": setting})
	return errr
}
