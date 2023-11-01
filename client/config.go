package client

import (
	"fmt"
	"net/url"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

type Config struct {
	Address   string `envconfig:"ADDRESS"`
	UserAgent string `envconfig:"USER_AGENT"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(loader ConfigLoader) error {
	if loader == nil {
		return fmt.Errorf("no ConfigLoader provided")
	}
	return loader.LoadClient(c)
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("address is missing")
	} else if _, err := url.Parse(c.Address); err != nil {
		return errors.New("address is invalid")
	}
	if c.UserAgent == "" {
		return errors.New("user agent is missing")
	}

	return nil
}

// ConfigLoader abstracts the method by which config values are loaded.
type ConfigLoader interface {
	// LoadClient sets config values for the properties of Config.
	LoadClient(*Config) error
}

// configReporterLoader adapts a config.Reporter to implement ConfigLoader.
type configReporterLoader struct {
	Reporter config.Reporter
}

func NewConfigReporterLoader(reporter config.Reporter) *configReporterLoader {
	return &configReporterLoader{
		Reporter: reporter,
	}
}

// LoadClient implements ConfigLoader.
func (l *configReporterLoader) LoadClient(cfg *Config) error {
	cfg.Address = l.Reporter.GetWithDefault("address", cfg.Address)
	cfg.UserAgent = l.Reporter.GetWithDefault("user_agent", cfg.UserAgent)
	return nil
}

// envconfigLoader adapts envconfig to implement ConfigLoader.
type envconfigLoader struct {
	Prefix string
}

func NewEnvconfigLoader(prefix string) *envconfigLoader {
	return &envconfigLoader{
		Prefix: prefix,
	}
}

// LoadClient implements ConfigLoader.
func (l *envconfigLoader) LoadClient(cfg *Config) error {
	return envconfig.Process(l.Prefix, cfg)
}
