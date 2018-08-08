package config

import "github.com/caarlos0/env"

type Configuration struct {
	Port     int    `env:"PORT" envDefault:"8080"`
	Hostname string `env:"HOSTNAME" envDefault:"localhost:8080"`
}

func ParseConfiguration() (Configuration, error) {
	config := Configuration{}
	if err := env.Parse(&config); err != nil {
		return config, err
	}
	return config, nil
}
