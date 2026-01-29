package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type HttpConfig struct {
	Port int
}
type Config struct {
	ENV      string
	GRPC     GRPCConfig
	Postgres PostgresConfig
	JWT      JWTConfig
	HTTP     HttpConfig
	Redis    RedisConfig
}

type RedisConfig struct {
	Port string
	Host string
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
	SSLMode  string
	Url      string
}

type JWTConfig struct {
	Secret   string        `env:"JWT_SECRET"`
	TokenTTL time.Duration `env:"JWT_TOKEN_TTL"`
}
type PathDB struct {
	path string
}

func InitConfig() error {
	// 1. Настраиваем автоматическое чтение переменных окружения (для Docker)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 2. Настраиваем чтение из файла (для локальной разработки)
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// 3. Пытаемся прочитать файл
	if err := viper.ReadInConfig(); err != nil {
		// Если файла нет — это НЕ ошибка, так как мы в Docker
		// и переменные придут через AutomaticEnv().
		if os.IsNotExist(err) {
			return nil
		}

		// Проверка на специфичную ошибку Viper (на всякий случай)
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil
		}
		// Если файл есть, но он битый (синтаксис) — возвращаем ошибку
		return err
	}

	return nil
}

func GetConfig() *Config {
	return &Config{
		ENV: viper.GetString("ENV"),
		GRPC: GRPCConfig{
			Port:    viper.GetInt("GRPC_PORT"),
			Timeout: viper.GetDuration("GRPC_TIMEOUT"),
		},
		HTTP: HttpConfig{
			Port: viper.GetInt("HTTP_PORT"),
		},
		Postgres: PostgresConfig{
			Host:     viper.GetString("POSTGRES_HOST"),
			Port:     viper.GetInt("POSTGRES_EXTERNAL_PORT"),
			DB:       viper.GetString("POSTGRES_DB"),
			UserName: viper.GetString("POSTGRES_USER"),
			Password: viper.GetString("POSTGRES_PASSWORD"),
			SSLMode:  viper.GetString("POSTGRES_SSL_MODE"),
			Url:      viper.GetString("POSTGRES_URI"),
		},
		Redis: RedisConfig{
			Host: viper.GetString("REDIS_HOST"),
			Port: viper.GetString("REDIS_PORT"),
		},
		JWT: JWTConfig{
			Secret:   viper.GetString("JWT_SECRET"),
			TokenTTL: viper.GetDuration("JWT_TOKEN_TTL"),
		},
	}
}

func (c *Config) GetPostgresPath() (connectionString string) {
	cfg := c.Postgres
	if cfg.Url != "" {
		connectionString = cfg.Url
	} else {
		connectionString = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.UserName, cfg.Password, cfg.DB, cfg.SSLMode)
	}
	return
}
