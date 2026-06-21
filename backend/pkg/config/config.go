package config

import (
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

var (
	cfg Config
)

type (
	Config struct {
		ConnectionString string `mapstructure:"connection_string"`
	}
)

func New() {
	config.WithOptions(config.ParseEnv)
	config.AddDriver(yaml.Driver)
	if err := config.LoadFiles("config/dev.yaml"); err != nil {
		panic(err)
	}
	//fmt.Printf("config data: \n %#v\n", config.Data()["db-ws"])
	if err := config.Decode(&cfg); err != nil {
		panic(err)
	}
}

func Get() *Config {
	return &cfg
}
