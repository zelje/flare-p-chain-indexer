package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	Host     string `yaml:"host" envconfig:"DB_HOST"`
	Port     int    `yaml:"port" envconfig:"DB_PORT"`
	Database string `yaml:"database" envconfig:"DB_DATABASE"`
	Username string `yaml:"username" envconfig:"DB_USERNAME"`
	Password string `yaml:"password" envconfig:"DB_PASSWORD"`
}

func ParseConfigFile(cfg interface{}, fileName string) error {
	f, err := os.Open(fileName)
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

func ReadEnv(cfg interface{}) error {
	err := envconfig.Process("", cfg)
	if err != nil {
		return fmt.Errorf("error reading env config: %w", err)
	}
	return nil
}
