package config

type Config struct {
	Server struct {
		Port         int    `mapstructure:"port"`
		Domain       string `mapstructure:"domain"`
		HTTPS        bool   `mapstructure:"https" default:"false"`
		EnableGinLog bool   `mapstructure:"enable_gin_log" default:"true"`
		LogLevel     string `mapstructure:"log_level" default:"INFO"`
	} `mapstructure:"server"`
	Frontend struct {
		Domain string `mapstructure:"domain"`
		Port   int    `mapstructure:"port"`
	} `mapstructure:"frontend"`
	MongoDB struct {
		URI      string `mapstructure:"uri"`
		Database string `mapstructure:"database"`
	} `mapstructure:"mongodb"`
	JWT struct {
		SecretKey string `mapstructure:"secretKey"`
	} `mapstructure:"JWT"`
	Discord struct {
		ClientID     string `mapstructure:"clientId"`
		ClientSecret string `mapstructure:"clientSecret"`
		AdminID      string `mapstructure:"adminId"`
		GuildID      string `mapstructure:"guildId"`
	} `mapstructure:"discord"`
	Redis struct {
		URI string `mapstructure:"uri"`
	} `mapstructure:"redis"`
	RateLimit struct {
		Enabled bool `json:"enabled"`

		// 登录端点 (IP 维度)
		LoginPerMinute     int64 `json:"login_per_minute"`
		OAuthInitPerMinute int64 `json:"oauth_init_per_minute"`
		LogoutPerMinute    int64 `json:"logout_per_minute"`

		// 聊天端点 (用户ID 维度)
		ChatPostPerMinute   int64 `json:"chat_post_per_minute"`
		ChatDeletePerMinute int64 `json:"chat_delete_per_minute"`

		// 账户管理
		ChangePasswordPerHour int64 `json:"change_password_per_hour"`
	}
}
