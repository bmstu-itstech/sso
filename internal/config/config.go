package config

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	ENV      string
	GRPC     GRPCConfig
	Postgres PostgresConfig
	JWT      JWTConfig
}

type GRPCConfig struct {
	Port    int
	Timeout time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     int
	DB       string
	UserName string
	Password string
}

type JWTConfig struct {
	Secret   string        `env:"JWT_SECRET"`
	TokenTTL time.Duration `env:"JWT_TOKEN_TTL"`
}

func InitConfig() error {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	return viper.ReadInConfig()
}

func GetConfig() *Config {
	return &Config{
		ENV: viper.GetString("ENV"),
		GRPC: GRPCConfig{
			Port:    viper.GetInt("GRPC_PORT"),
			Timeout: viper.GetDuration("GRPC_TIMEOUT"),
		},
		Postgres: PostgresConfig{
			Host:     viper.GetString("POSTGRES_HOST"),
			Port:     viper.GetInt("POSTGRES_PORT"),
			DB:       viper.GetString("POSTGRES_DB"),
			UserName: viper.GetString("POSTGRES_USER"),
			Password: viper.GetString("POSTGRES_PASSWORD"),
		},
		JWT: JWTConfig{
			Secret:   viper.GetString("JWT_SECRET"),
			TokenTTL: viper.GetDuration("JWT_TOKEN_TTL"),
		},
	}
}
