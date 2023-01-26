package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DB struct {
		Host     string `yaml:"host" envconfig:"DB_HOST"`
		Port     int    `yaml:"port" envconfig:"DB_PORT"`
		Database string `yaml:"database" envconfig:"DB_DATABASE"`
		Username string `yaml:"username" envconfig:"DB_USERNAME"`
		Password string `yaml:"password" envconfig:"DB_PASSWORD"`
	} `yaml:"db"`
	Chain struct {
		IndexerURL string `yaml:"indexer_url" envconfig:"CHAIN_INDEXER_URL"`
	} `yaml:"chain"`
}

func BuildConfig() (*Config, error) {
	cfg := Config{}
	err := parseConfigFile(&cfg)
	if err != nil {
		return nil, err
	}
	err = readEnv(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func parseConfigFile(cfg *Config) error {
	f, err := os.Open("config.yml")
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}
	return nil
}

func readEnv(cfg *Config) error {
	err := envconfig.Process("", cfg)
	if err != nil {
		return fmt.Errorf("error reding env config: %w", err)
	}
	return nil
}
