package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
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
	// Настройка чтения переменных окружения
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Чтение .env файла (если существует)
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	// Не возвращаем ошибку, если файл не найден (переменных окружения могут быть достаточны)
	_ = viper.ReadInConfig()

	return nil
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
			Port:     viper.GetInt("POSTGRES_EXTERNAL_PORT"),
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
