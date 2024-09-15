package config
type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
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
	Domain string `mapstructure:"domain"`
	HTTPS  bool   `mapstructure:"https" default:"false"`
}