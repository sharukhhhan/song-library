package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"path"
	"time"
)

type (
	Config struct {
		HTTP        `yaml:"http"`
		Log         `yaml:"log"`
		PG          `yaml:"postgres"`
		ExternalAPI `yaml:"externalAPI"`
	}

	HTTP struct {
		Port            string        `env-required:"false" yaml:"port"`
		ShutdownTimeout time.Duration `env-required:"false" yaml-default:"5s" yaml:"shutdownTimeout"`
	}

	Log struct {
		Level string `env-required:"false" yaml:"level"`
	}

	PG struct {
		URL           string `env-required:"false" env:"PG_URL"`
		MigrationPath string `env-required:"false" env-default:"./migrations" env:"PG_MIGRATION_PATH"`
	}

	ExternalAPI struct {
		URL string `env-required:"false" yaml:"URL"`
	}
)

func NewConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	err := cleanenv.ReadConfig(path.Join("./", configPath), cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading env: %w", err)
	}

	return cfg, nil
}
