package config

import (
	"github.com/spf13/viper"

	"url-shortner/log"
)

type Config struct {
	Database SQLDatabase `yaml:"database"`
}

type SQLDatabase struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DB       string `yaml:"db"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Retry    int    `yaml:"retry"`
}

func Init() Config {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("err: %s", err)
	}

	cfg := new(Config)
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("err: %s", err)
	}

	return *cfg
}
