package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
)

const (
	LOCAL_CONFIG_FILE string = "config.local.toml"
	CONFIG_FILE       string = "config.toml"
)

var (
	GlobalConfigCallback ConfigCallback = ConfigCallback{}
)

type GlobalConfig interface {
	AddressHRP() string
	LoggerConfig() LoggerConfig
}

type LoggerLevel string

type LoggerConfig struct {
	Level       string `toml:"level"` // valid values are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL (zap)
	File        string `toml:"file"`
	MaxFileSize int    `toml:"max_file_size"` // In megabytes
	Console     bool   `toml:"console"`
}

type DBConfig struct {
	Host     string `toml:"host" envconfig:"DB_HOST"`
	Port     int    `toml:"port" envconfig:"DB_PORT"`
	Database string `toml:"database" envconfig:"DB_DATABASE"`
	Username string `toml:"username" envconfig:"DB_USERNAME"`
	Password string `toml:"password" envconfig:"DB_PASSWORD"`
}

type GlobalConfigCallbackFunc func(GlobalConfig)

type ConfigCallback struct {
	globalCallbacks []GlobalConfigCallbackFunc
}

func (cc *ConfigCallback) AddCallback(f GlobalConfigCallbackFunc) {
	cc.globalCallbacks = append(cc.globalCallbacks, f)
}

func (cc *ConfigCallback) Call(config GlobalConfig) {
	for _, gc := range cc.globalCallbacks {
		gc(config)
	}
}

func ParseConfigFile(cfg interface{}, fileName string, allowMissing bool) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		if allowMissing {
			return nil
		} else {
			return fmt.Errorf("error opening config file: %w", err)
		}
	}

	_, err = toml.Decode(string(content), cfg)
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
