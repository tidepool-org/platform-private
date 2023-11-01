package platform

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/config"
)

// Config extends client.Config with a service secret.
type Config struct {
	*client.Config
	ServiceSecret string `envconfig:"SERVICE_SECRET"`
}

func NewConfig() *Config {
	return &Config{
		Config: client.NewConfig(),
	}
}

type ConfigLoader interface {
	client.ConfigLoader
	LoadPlatform(*Config) error
}

func (c *Config) Load(loader ConfigLoader) error {
	if loader == nil {
		return fmt.Errorf("no ConfigLoader provided")
	}
	return loader.LoadPlatform(c)
}

// configReporterLoader adapts config.Reporter to implement ConfigLoader.
type configReporterLoader struct {
	Reporter config.Reporter
	client.ConfigLoader
}

func NewConfigReporterLoader(reporter config.Reporter) *configReporterLoader {
	return &configReporterLoader{
		ConfigLoader: client.NewConfigReporterLoader(reporter),
		Reporter:     reporter,
	}
}

// LoadPlatform implements ConfigLoader.
func (l *configReporterLoader) LoadPlatform(cfg *Config) error {
	if err := l.ConfigLoader.LoadClient(cfg.Config); err != nil {
		return err
	}
	cfg.ServiceSecret = l.Reporter.GetWithDefault("service_secret", cfg.ServiceSecret)
	return nil
}

// envconfigLoader adapts envconfig to implement ConfigLoader.
type envconfigLoader struct {
	Prefix       string
	ConfigLoader client.ConfigLoader
}

func NewEnvconfigLoader(prefix string) *envconfigLoader {
	return &envconfigLoader{
		Prefix:       prefix,
		ConfigLoader: client.NewEnvconfigLoader(prefix),
	}
}

// LoadPlatform implements ConfigLoader.
func (l *envconfigLoader) LoadPlatform(cfg *Config) error {
	if err := l.ConfigLoader.LoadClient(cfg.Config); err != nil {
		return err
	}
	return envconfig.Process(l.Prefix, cfg)
}
