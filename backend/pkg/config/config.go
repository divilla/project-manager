package config

import (
	"os"
	"strings"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

var (
	cfg Config
)

type (
	Config struct {
		ConnectionString string `mapstructure:"connection_string"`
		Port             string `mapstructure:"port"`
		CORSOrigins      string `mapstructure:"cors_origins"`
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

	applyDefaults()
	applyEnv()
}

func Get() *Config {
	return &cfg
}

func applyDefaults() {
	if cfg.ConnectionString == "" {
		cfg.ConnectionString = "postgresql://localhost:5432/postgres"
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.CORSOrigins == "" {
		cfg.CORSOrigins = "http://localhost:8000"
	}
}

func applyEnv() {
	if value := os.Getenv("DATABASE_URL"); value != "" {
		cfg.ConnectionString = value
	}
	if value := os.Getenv("PORT"); value != "" {
		cfg.Port = value
	}
	if value := os.Getenv("CORS_ORIGINS"); value != "" {
		cfg.CORSOrigins = value
	}
}

func (c Config) Addr() string {
	if strings.HasPrefix(c.Port, ":") {
		return c.Port
	}
	return ":" + c.Port
}

func (c Config) AllowedOrigins() []string {
	var origins []string
	for _, origin := range strings.Split(c.CORSOrigins, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
