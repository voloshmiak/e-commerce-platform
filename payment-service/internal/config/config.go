package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DB struct {
		Host           string `env:"DB_HOST" envDefault:"postgres-payment"`
		Port           string `env:"DB_PORT" envDefault:"5434"`
		User           string `env:"DB_USER" envDefault:"postgres"`
		Password       string `env:"DB_PASSWORD" envDefault:"password"`
		Name           string `env:"DB_NAME" envDefault:"paymentdb"`
		MigrationsPath string `env:"DB_MIGRATIONS_PATH" envDefault:"./migrations"`
	}
	Server struct {
		Port string `env:"SERVER_PORT" envDefault:":8080"`
	}
	CartClient struct {
		Host string `env:"CART_HOST" envDefault:"shopping-cart-service"`
		Port string `env:"CART_PORT" envDefault:"8080"`
	}
	Kafka struct {
		Host string `env:"KAFKA_HOST" envDefault:"kafka"`
		Port string `env:"KAFKA_PORT" envDefault:"9092"`
	}
	Stripe struct {
		SecretKey string `env:"STRIPE_SECRET" envDefault:"your_stripe_secret_key"`
	}
}

func New() (*Config, error) {
	var config Config

	if err := cleanenv.ReadEnv(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
