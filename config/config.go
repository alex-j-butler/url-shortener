package config

import (
	"errors"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config is a struct representing configuration options.
type Config struct {
	PostgresDSN string `yaml:"postgres_dsn"`

	RedisAddress  string `yaml:"redis_address"`
	RedisPassword string `yaml:"redis_password"`
	RedisDB       int    `yaml:"redis_db"`

	BindAddress string `yaml:"bind_address"`
	BindPort    int    `yaml:"bind_port"`

	DefaultURL string `yaml:"default_url"`
	BaseURL    string `yaml:"base_url"`
}

var (
	errConfigNotFound = errors.New("Configuration file not found")
	errConfigParse    = errors.New("Configuration file could not be parsed")
)

// Conf is the global configuration struct.
var Conf Config

// PopulateConfig loads the configuration from the 'config.yml' file and
// parses it into the global configuration struct.
func PopulateConfig() error {
	configBytes, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		return errConfigNotFound
	}
	err = yaml.Unmarshal(configBytes, &Conf)
	if err != nil {
		return errConfigParse
	}

	return nil
}
