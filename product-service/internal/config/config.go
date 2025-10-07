package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Mongo struct {
		User     string `env:"MONGO_USERNAME" envDefault:"root"`
		Password string `env:"MONGO_PASSWORD" envDefault:"password"`
		Host     string `env:"MONGO_HOST" envDefault:"mongodb"`
		Port     string `env:"MONGO_PORT" envDefault:"27017"`
		DBName   string `env:"MONGO_DB_NAME" envDefault:"ecommerce"`
	}
	Elastic struct {
		Host string `env:"ELASTIC_HOST" envDefault:"elasticsearch"`
		Port string `env:"ELASTIC_PORT" envDefault:"9200"`
	}
	Server struct {
		Port string `env:"SERVER_PORT" envDefault:":8080"`
	}
	Kafka struct {
		Host string `env:"KAFKA_HOST" envDefault:"kafka"`
		Port string `env:"KAFKA_PORT" envDefault:"9092"`
	}
}

func New() (*Config, error) {
	var config Config

	if err := cleanenv.ReadEnv(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
