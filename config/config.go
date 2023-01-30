package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DB      DBConfig      `yaml:"db"`
	Chain   ChainConfig   `yaml:"chain"`
	Indexer IndexerConfig `yaml:"indexer"`
}

type DBConfig struct {
	Host     string `yaml:"host" envconfig:"DB_HOST"`
	Port     int    `yaml:"port" envconfig:"DB_PORT"`
	Database string `yaml:"database" envconfig:"DB_DATABASE"`
	Username string `yaml:"username" envconfig:"DB_USERNAME"`
	Password string `yaml:"password" envconfig:"DB_PASSWORD"`
}

type IndexerConfig struct {
	TimeoutMillis int `yaml:"timeout_millis"`
	BatchSize     int `yaml:"batch_size"`
}

type ChainConfig struct {
	IndexerURL string `yaml:"indexer_url" envconfig:"CHAIN_INDEXER_URL"`
}

func newConfig() *Config {
	return &Config{
		Indexer: IndexerConfig{
			TimeoutMillis: 3000,
			BatchSize:     10,
		},
	}
}

func BuildConfig() (*Config, error) {
	cfg := newConfig()
	err := parseConfigFile(cfg)
	if err != nil {
		return nil, err
	}
	err = readEnv(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
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
