// Go-Service/src/main/infrastructure/initializer/initializer.go
package initializer

import (
	"Go-Service/src/main/infrastructure/config"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

func InitConfig() {
	config.LoadConfig()
}

func InitMongoClient() {
	uri := config.AppConfig.MongoDB.URI
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB!")
	Client = client
	DB = client.Database(config.AppConfig.MongoDB.Database)
}

func CleanupMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	DB.Collection("skeletons").Drop(ctx)
	Client.Disconnect(ctx)
}
