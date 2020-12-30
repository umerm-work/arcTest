package config

import (
	"github.com/spf13/viper"
)

// Config describes service config
// All values are located in ENV variables
// or have defaults
type Config struct {
	ApplicationName   string
	AppPort           string
	DBHost            string
	DBUser            string
	DBPasswd          string
	DbName            string
	DbUsersCollection string
	DbIdeasCollection string
}

// ReadEnv returns environment config
func ReadEnv() Config {

	viper.AutomaticEnv()

	viper.SetEnvPrefix("APP")

	viper.SetDefault("DB_HOST", "172.17.0.3:27017")
	viper.SetDefault("DB_USER", "")
	viper.SetDefault("DB_PASSWORD", "")
	viper.SetDefault("DB_NAME", "arc")
	viper.SetDefault("DB_USER_COLLECTION", "users")
	viper.SetDefault("DB_IDEA_COLLECTION", "ideas")
	viper.SetDefault("APP_PORT", "8000")
	viper.SetDefault("APPLICATION_NAME", "are-test")

	Settings := Config{
		DBHost:            viper.GetString("DB_HOST"),
		DBUser:            viper.GetString("DB_USER"),
		DBPasswd:          viper.GetString("DB_PASSWORD"),
		DbName:            viper.GetString("DB_NAME"),
		ApplicationName:   viper.GetString("APPLICATION_NAME"),
		AppPort:           viper.GetString("APP_PORT"),
		DbUsersCollection: viper.GetString("DB_USER_COLLECTION"),
		DbIdeasCollection: viper.GetString("DB_IDEA_COLLECTION"),
	}

	return Settings
}
