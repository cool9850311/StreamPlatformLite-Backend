// Go-Service/src/main/infrastructure/initializer/initializer.go
package initializer

import (
	"Go-Service/src/main/application/interface/stream"
	"Go-Service/src/main/application/usecase"
	domainLogger "Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/cache"
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/livestream"
	infraLogger "Go-Service/src/main/infrastructure/logger"
	"Go-Service/src/main/infrastructure/repository"
	"Go-Service/src/main/infrastructure/util"
	"context"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	migrations "Go-Service/migrations"
)

var GormDB *gorm.DB
var Log domainLogger.Logger
var LiveStreamService stream.ILivestreamService
var RedisClient *redis.Client
var cronJob *cron.Cron

func InitLog() {
	var err error
	logLevel := config.AppConfig.Server.LogLevel
	if logLevel == "" {
		logLevel = "INFO"
	}
	Log, err = infraLogger.NewLogger("application.log", logLevel)
	if err != nil {
		panic(err)
	}
}
func InitConfig() {
	config.LoadConfig()
}

func InitSchema() {
	if !config.AppConfig.PostgreSQL.AutoMigrateSchema {
		log.Println("[schema] Auto-migrate disabled, skipping")
		return
	}
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatalf("failed to create migration source: %v", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, config.AppConfig.PostgreSQL.DSN)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("schema migration failed: %v", err)
	}
	log.Println("[schema] Schema migration done.")
}

func InitPostgresClient() {
	dsn := config.AppConfig.PostgreSQL.DSN
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}
	GormDB = db
	log.Println("Connected to PostgreSQL")
}

func InitRedisClient() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.Redis.URI,
		Password: "",
		DB:       0,
	})
}

func InitLiveStreamService(log domainLogger.Logger, db *gorm.DB) {
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
	livestreamRepo := repository.NewPostgresLivestreamRepository(db)
	result, err := livestreamRepo.GetOne()
	if err != nil {
		log.Info(context.TODO(), "No Stream Found: "+err.Error())
		return
	}
	LiveStreamService.OpenStream(result.Name, result.UUID, result.APIKey, result.IsRecord)
	log.Info(context.TODO(), "Livestream Started: "+result.UUID)
	// time.Sleep(time.Second)

	// // Open a stream
	// LiveStreamService.OpenStream("test1", "test2", "test")
	// LiveStreamService.CloseStream("test2")
}
func InitCronJob(log domainLogger.Logger, db *gorm.DB) {
	cronJob = cron.New()
	viewerCountCache := cache.NewRedisViewerCount(RedisClient)
	chatCache := cache.NewRedisChat(RedisClient)
	fileCache := cache.NewFileCache()
	ffmpegLibrary := util.NewFfmpegLibrary()
	livestreamRepo := repository.NewPostgresLivestreamRepository(db)
	livestreamUseCase := usecase.NewLivestreamUsecase(livestreamRepo, log, config.AppConfig, LiveStreamService, viewerCountCache, chatCache, fileCache, ffmpegLibrary)
	cronJob.AddFunc("@every 10s", func() {
		log.Info(context.Background(), "Running viewer count cleanup")
		ls, err := livestreamRepo.GetOne()
		if err != nil {
			log.Error(context.Background(), "Error fetching livestream: "+err.Error())
			return
		}
		livestreamUseCase.RemoveViewerCount(context.Background(), ls.UUID, 10)
	})

	cronJob.Start()
}
