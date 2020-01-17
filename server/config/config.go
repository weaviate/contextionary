package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Config is used to load application wide config from the environment
type Config struct {
	logger        logrus.FieldLogger
	KNNFile       string
	IDXFile       string
	StopwordsFile string

	SchemaProviderURL string
	SchemaProviderKey string
	ExtensionsPrefix  string

	ServerPort int

	OccurrenceWeightStrategy     string
	OccurrenceWeightLinearFactor float32
	MaxCompoundWordLength        int

	LogLevel string
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

	sw, err := c.requiredString("STOPWORDS_FILE")
	if err != nil {
		return err
	}
	c.StopwordsFile = sw

	sp, err := c.requiredString("SCHEMA_PROVIDER_URL")
	if err != nil {
		return err
	}
	c.SchemaProviderURL = sp

	spk := c.optionalString("SCHEMA_PROVIDER_KEY", "/weaviate/schema/state")
	c.SchemaProviderKey = spk

	ep := c.optionalString("EXTENSIONS_PREFIX", "/contextionary/")
	c.ExtensionsPrefix = ep

	port, err := c.optionalInt("SERVER_PORT", 9999)
	if err != nil {
		return err
	}
	c.ServerPort = port

	factor, err := c.optionalFloat32("OCCURRENCE_WEIGHT_LINEAR_FACTOR", 0.5)
	if err != nil {
		return err
	}
	c.OccurrenceWeightLinearFactor = factor

	strategy := c.optionalString("OCCURRENCE_WEIGHT_STRATEGY", "log")
	c.OccurrenceWeightStrategy = strategy

	// this should match the underlying vector db file, a smaller value than in
	// the vector file will lead to missing out on compound words, whereas a
	// larger value will lead to unnecessary lookups slowing down the
	// vectorization process
	compoundLength, err := c.optionalInt("MAX_COMPOUND_WORD_LENGTH", 4)
	if err != nil {
		return err
	}
	c.MaxCompoundWordLength = compoundLength

	loglevel := c.optionalString("LOG_LEVEL", "info")
	c.LogLevel = loglevel

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

func (c *Config) optionalFloat32(varName string, defaultValue float32) (float32, error) {
	value := os.Getenv(varName)
	if value == "" {
		c.logger.Infof("optional var '%s' is not set, defaulting to '%v'",
			varName, defaultValue)
		return defaultValue, nil
	}

	asFloat, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0, fmt.Errorf("cannot convert value of var '%s' ('%v') to int: %s",
			varName, value, err)
	}

	return float32(asFloat), nil
}

func (c *Config) requiredString(varName string) (string, error) {
	value := os.Getenv(varName)
	if value == "" {
		return "", fmt.Errorf("required variable '%s' is not set", varName)
	}

	return value, nil
}

func (c *Config) optionalString(varName, defaultInput string) string {
	value := os.Getenv(varName)
	if value == "" {
		c.logger.Infof("optional var '%s' is not set, defaulting to '%v'",
			varName, defaultInput)
		return defaultInput
	}

	return value
}
