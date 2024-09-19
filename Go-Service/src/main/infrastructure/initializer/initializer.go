// Go-Service/src/main/infrastructure/initializer/initializer.go
package initializer

import (
	domainLogger "Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	infraLogger "Go-Service/src/main/infrastructure/logger"
	"context"
	"log"
	"time"
	"github.com/robfig/cron/v3"
	"Go-Service/src/main/application/interface/stream"
	"Go-Service/src/main/infrastructure/livestream"
	"Go-Service/src/main/infrastructure/repository"
	"Go-Service/src/main/infrastructure/cache"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"Go-Service/src/main/application/usecase"
)

var Client *mongo.Client
var DB *mongo.Database
var Log domainLogger.Logger
var LiveStreamService stream.ILivestreamService
var RedisClient *redis.Client
var cronJob *cron.Cron
func InitLog() {
	var err error
	Log, err = infraLogger.NewLogger("application.log")
	if err != nil {
		panic(err)
	}
}
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

	// Create a unique index on the username field
	userCollection := DB.Collection("users")
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = userCollection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	log.Println("Created unique index on username field")
}

func InitRedisClient() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.Redis.URI,
		Password: "",
		DB:       0,
	})
}

func CleanupMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	DB.Collection("skeletons").Drop(ctx)
	DB.Collection("users").Drop(ctx)
	Client.Disconnect(ctx)
}

func InitLiveStreamService(log domainLogger.Logger, db *mongo.Database) {
	LiveStreamService = livestream.NewLivestreamService(log)

	// Start the service
	err := LiveStreamService.StartService()
	if err != nil {
		log.Fatal(context.TODO(), "Failed to start LiveStreamService: "+err.Error())
	}

	// Run the service loop in a goroutine
	go func() {
		err := LiveStreamService.RunLoop()
		if err != nil {
			log.Fatal(context.TODO(), "RunLoop error: "+err.Error())
		}
	}()
	livestreamRepo := repository.NewMongoLivestreamRepository(db)
	result, err := livestreamRepo.GetOne()
	if err != nil {
		log.Info(context.TODO(), "No Stream Found"+err.Error())
		return
	}
	LiveStreamService.OpenStream(result.Name, result.UUID, result.APIKey, result.OutputPathUUID)
	log.Info(context.TODO(), "Livestream Started: "+result.UUID)
	// time.Sleep(time.Second)

	// // Open a stream
	// LiveStreamService.OpenStream("test1", "test2", "test")
	// LiveStreamService.CloseStream("test2")
}
func InitCronJob(log domainLogger.Logger, db *mongo.Database) {
	cronJob = cron.New()
	viewerCountCache := cache.NewRedisViewerCount(RedisClient)
	chatCache := cache.NewRedisChat(RedisClient)
	livestreamRepo := repository.NewMongoLivestreamRepository(db)
	livestreamUseCase := usecase.NewLivestreamUsecase(livestreamRepo, log, config.AppConfig, LiveStreamService, viewerCountCache, chatCache)
	cronJob.AddFunc("@every 10s", func() {
		log.Info(context.Background(), "Running viewer count cleanup")
		uuid,err := livestreamRepo.GetOne()
		if err != nil {
			log.Error(context.Background(), "Error fetching livestream: "+err.Error())
			return
		}
		livestreamUseCase.RemoveViewerCount(context.Background(), uuid.UUID, 10)
	})
	
	cronJob.Start()
}