package repository

import (
	"Go-Service/src/main/application/interface/repository"
	"Go-Service/src/main/domain/entity/account"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"Go-Service/src/main/domain/entity/errors"
)

type MongoAccountRepository struct {
	collection *mongo.Collection
}

func NewMongoAccountRepository(db *mongo.Database) repository.AccountRepository {
	collection := db.Collection("accounts")

	// Create a unique index on the username field
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"username", 1}}, // 1 for ascending order
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		// Handle the error according to your application's needs
		log.Fatalf("Failed to create index: %v", err)
	}

	return &MongoAccountRepository{
		collection: collection,
	}
}

func (r *MongoAccountRepository) Create(acc account.Account) error {
	
	_, err := r.collection.InsertOne(context.TODO(), acc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.ErrDuplicate
		}
		return err
	}
	return err
}

func (r *MongoAccountRepository) GetAll() ([]account.Account, error) {
	var accounts []account.Account
	cursor, err := r.collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var acc account.Account
		if err := cursor.Decode(&acc); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *MongoAccountRepository) GetByUsername(username string) (*account.Account, error) {
	var acc account.Account
	filter := bson.D{{"username", username}}
	err := r.collection.FindOne(context.TODO(), filter).Decode(&acc)
	if err == mongo.ErrNoDocuments {
		return nil, errors.ErrNotFound
	}
	return &acc, err
}

func (r *MongoAccountRepository) Update(acc account.Account) error {
	filter := bson.D{{"username", acc.Username}}
	update := bson.D{{"$set", acc}}
	_, err := r.collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *MongoAccountRepository) Delete(username string) error {
	filter := bson.D{{"username", username}}
	_, err := r.collection.DeleteOne(context.TODO(), filter)
	return err
}
