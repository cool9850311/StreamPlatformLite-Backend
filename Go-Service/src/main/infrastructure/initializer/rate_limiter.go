package initializer

import (
	"Go-Service/src/main/infrastructure/config"
	"context"
	"time"

	"github.com/ulule/limiter/v3"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

var (
	// User-based limiters
	ChatPostLimiter   *limiter.Limiter
	ChatDeleteLimiter *limiter.Limiter
)

func InitRateLimiters() {
	if !config.AppConfig.RateLimit.Enabled {
		Log.Info(context.TODO(), "Rate limiting is disabled")
		return
	}

	store, err := sredis.NewStoreWithOptions(RedisClient, limiter.StoreOptions{
		Prefix: "rate_limit:",
	})
	if err != nil {
		Log.Fatal(context.TODO(), "Failed to create Redis store for rate limiter: "+err.Error())
	}

	// User-based limiters
	ChatPostLimiter = limiter.New(store, limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  config.AppConfig.RateLimit.ChatPostPerMinute,
	})

	ChatDeleteLimiter = limiter.New(store, limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  config.AppConfig.RateLimit.ChatDeletePerMinute,
	})

	Log.Info(context.TODO(), "Rate limiters initialized successfully")
}
