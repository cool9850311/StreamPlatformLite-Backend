package initializer

import (
	"Go-Service/src/main/infrastructure/config"
	"context"
	"time"

	"github.com/ulule/limiter/v3"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

var (
	// IP-based limiters
	LoginLimiter     *limiter.Limiter
	OAuthInitLimiter *limiter.Limiter
	LogoutLimiter    *limiter.Limiter

	// User-based limiters
	ChatPostLimiter       *limiter.Limiter
	ChatDeleteLimiter     *limiter.Limiter
	ChangePasswordLimiter *limiter.Limiter
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

	// IP-based limiters
	LoginLimiter = limiter.New(store, limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  config.AppConfig.RateLimit.LoginPerMinute,
	})

	OAuthInitLimiter = limiter.New(store, limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  config.AppConfig.RateLimit.OAuthInitPerMinute,
	})

	LogoutLimiter = limiter.New(store, limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  config.AppConfig.RateLimit.LogoutPerMinute,
	})

	// User-based limiters
	ChatPostLimiter = limiter.New(store, limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  config.AppConfig.RateLimit.ChatPostPerMinute,
	})

	ChatDeleteLimiter = limiter.New(store, limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  config.AppConfig.RateLimit.ChatDeletePerMinute,
	})

	ChangePasswordLimiter = limiter.New(store, limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  config.AppConfig.RateLimit.ChangePasswordPerHour,
	})

	Log.Info(context.TODO(), "Rate limiters initialized successfully")
}
