package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Config
	AppSecret string
}
type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
	Clients     ClientsConfig `yaml:"clients" env-required:"true"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Client struct {
	Address string `yaml:"address" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"4s"`
	Retries int `yaml:"retries" env-default:"3"`
	Insecure bool `yaml:"insecure" env-default:"false"`
	AppId int `yaml:"app_id" env-required:"true"`
}

type ClientsConfig struct {
	SSO Client `yaml:"sso" env-required:"true"`
}

func MustLoad() *AppConfig {
	configPath := os.Getenv("CONFIG_PATH")
	appSecret := os.Getenv("HTTP_SERVER_PASSWORD")
	if appSecret == "" {
		log.Fatal("HTTP_SERVER_PASSWORD is not set")
	}
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	appCfg := &AppConfig{
		AppSecret: appSecret,
		Config: cfg,
	}

	return appCfg
}
