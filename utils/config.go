package utils

import (
	"time"

	"github.com/spf13/viper"
)

// Config stores all configurations of the application load by viper
type Config struct {
	Environment			string				`mapstructure:"ENVIRONMENT"`
	DBDriver 			string				`mapstructure:"DB_DRIVER"`
	DSN 				string				`mapstructure:"DSN"`
	MigrationURL 		string				`mapstructure:"MIGRATION_URL"`
	HTTPServerAddress 	string 				`mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress 	string 				`mapstructure:"GRPC_SERVER_ADDRESS"`
	RedisAddress 		string 				`mapstructure:"REDIS_ADDRESS"`
	TokenSymmetryKey 	string				`mapstructure:"TOKEN_SYMMETRY_KEY"`
	AccessTokenDuration time.Duration 		`mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration 		`mapstructure:"REFRESH_TOKEN_DURATION"`
	EmailSenderName		string				`mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress	string				`mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword string				`mapstructure:"EMAIL_SENDER_PASSWORD"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}