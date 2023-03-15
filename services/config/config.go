package config

import (
	"flare-indexer/config"
)

type Config struct {
	DB       config.DBConfig     `toml:"db"`
	Logger   config.LoggerConfig `toml:"logger"`
	Services ServicesConfig      `toml:"services"`
}

type ServicesConfig struct {
	Address string `toml:"address"`
}

func newConfig() *Config {
	return &Config{
		Services: ServicesConfig{
			Address: "localhost:8000",
		},
	}
}

func (c Config) LoggerConfig() config.LoggerConfig {
	return c.Logger
}

func BuildConfig() (*Config, error) {
	cfg := newConfig()
	err := config.ParseConfigFile(cfg, config.CONFIG_FILE, false)
	if err != nil {
		return nil, err
	}
	err = config.ParseConfigFile(cfg, config.LOCAL_CONFIG_FILE, true)
	if err != nil {
		return nil, err
	}
	err = config.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
