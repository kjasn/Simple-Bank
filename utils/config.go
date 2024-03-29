package utils

import "github.com/spf13/viper"

// Config stores all configurations of the application load by viper
type Config struct {
	DBDriver 		string		`mapstructure:"DB_DRIVER"`
	DSN 			string		`mapstructure:"DSN"`
	ServerAddress 	string 		`mapstructure:"SERVER_ADDRESS"`
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