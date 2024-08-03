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
	// Load the URI from the config
	uri := config.AppConfig.MongoDB.URI

	// Create client options
	clientOptions := options.Client().ApplyURI(uri)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	// Context with timeout to use for ping and initial connection check
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ping the primary to verify connectivity
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB!")

	// Assign the client and database to global variables
	Client = client
	DB = client.Database(config.AppConfig.MongoDB.Database)
}

func CleanupMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	DB.Collection("skeletons").Drop(ctx)
	Client.Disconnect(ctx)
}
