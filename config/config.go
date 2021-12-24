package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig
	Client ClientConfig
}

type ServerConfig struct {
	BindAddr string
}

type ClientConfig struct {
	Prysm struct {
		Scheme string
		Host   string
		Port   string
	}
	Lighthouse struct {
		Scheme string
		Host   string
		Port   string
	}
	Teku struct {
		Scheme string
		Host   string
		Port   string
	}
	Nimbus struct {
		Scheme string
		Host   string
		Port   string
	}
	Geth struct {
		Scheme string
		Host   string
		Port   string
	}
	Erigon struct {
		Scheme string
		Host   string
		Port   string
	}
}

func NewConfig() (*Config, error) {
	// Set the file name of the configurations file
	viper.SetConfigName("config")
	// Set the path to look for the configurations file
	viper.AddConfigPath(".")
	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigType("yml")

	var config *Config

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
