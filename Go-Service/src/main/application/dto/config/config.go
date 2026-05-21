package config

type Config struct {
	Server struct {
		Port         int    `mapstructure:"port"`
		Domain       string `mapstructure:"domain"`
		RTMPHost     string `mapstructure:"rtmp_host"`
		HTTPS        bool   `mapstructure:"https" default:"false"`
		EnableGinLog bool   `mapstructure:"enable_gin_log" default:"true"`
		LogLevel     string `mapstructure:"log_level" default:"INFO"`
	} `mapstructure:"server"`
	Frontend struct {
		Domain string `mapstructure:"domain"`
		Port   int    `mapstructure:"port"`
	} `mapstructure:"frontend"`
	PostgreSQL struct {
		DSN               string `mapstructure:"dsn"`
		AutoMigrateSchema bool
	} `mapstructure:"postgresql"`
	JWT struct {
		SecretKey string `mapstructure:"secretKey"`
	} `mapstructure:"JWT"`
	Redis struct {
		URI string `mapstructure:"uri"`
	} `mapstructure:"redis"`
	RateLimit struct {
		Enabled bool `json:"enabled"`

		// 聊天端点 (用户ID 维度)
		ChatPostPerMinute   int64 `json:"chat_post_per_minute"`
		ChatDeletePerMinute int64 `json:"chat_delete_per_minute"`
	}
}
