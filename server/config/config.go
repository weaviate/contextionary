package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Config is used to load application wide config from the environment
type Config struct {
	logger  logrus.FieldLogger
	KNNFile string
	IDXFile string

	ServerPort int
}

// New Config from the environment. Errors if required env vars can't be found
func New(logger logrus.FieldLogger) (*Config, error) {
	cfg := &Config{logger: logger}
	if err := cfg.init(); err != nil {
		return nil, fmt.Errorf("could not load config from env: %v", err)
	}

	return cfg, nil
}

func (c *Config) init() error {
	knn, err := c.requiredString("KNN_FILE")
	if err != nil {
		return err
	}
	c.KNNFile = knn

	idx, err := c.requiredString("IDX_FILE")
	if err != nil {
		return err
	}
	c.IDXFile = idx

	port, err := c.optionalInt("SERVER_PORT", 9999)
	if err != nil {
		return err
	}
	c.ServerPort = port

	return nil
}

func (c *Config) optionalInt(varName string, defaultValue int) (int, error) {
	value := os.Getenv(varName)
	if value == "" {
		c.logger.Infof("optional var '%s' is not set, defaulting to '%v'",
			varName, defaultValue)
		return defaultValue, nil
	}

	asInt, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("cannot convert value of var '%s' ('%v') to int: %s",
			varName, value, err)
	}

	return asInt, nil
}

func (c *Config) requiredString(varName string) (string, error) {
	value := os.Getenv(varName)
	if value == "" {
		return "", fmt.Errorf("required variable '%s' is not set", varName)
	}

	return value, nil
}
