package config

import (
	// Built-in PKG
	"fmt"

	// External PKG
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		App    `yaml:"app"`
		DCMFM  `yaml:"dcmfm"`
		Log    `yaml:"logger"`
		Policy `yaml:"policy"`
	}

	// App -.
	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Type    string `env-required:"true" yaml:"type"    env:"APP_TYPE"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	// DCMFM -.
	DCMFM struct {
		AgentIP     string `env-required:"true" yaml:"agent_ip" env:"AGENT_IP"`
		AgentPort   string `env-required:"true" yaml:"agent_port" env:"AGENT_HTTP_PORT"`
		MonitorIP   string `env-required:"true" yaml:"monitor_ip" env:"MONITOR_IP"`
		MonitorPort string `env-required:"true" yaml:"monitor_port" env:"MONITOR_HTTP_PORT"`
		ManagerIP   string `env-required:"true" yaml:"manager_ip" env:"MANAGER_IP"`
		ManagerPort string `env-required:"true" yaml:"manager_port" env:"MANAGER_HTTP_PORT"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL"`
	}

	// Policy -.
	Policy struct {
		Alloc string `env-required:"true" yaml:"alloc" env:"ALLOC"`
		Free  string `env-required:"true" yaml:"free" env:"FREE"`
	}
)

func GetConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
