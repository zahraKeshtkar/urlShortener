package config

import (
	"github.com/spf13/viper"

	"url-shortner/log"
)

var defaultConfig *Config

type Config struct {
	Database    SQLDatabase   `yaml:"database"`
	HttpHandler HttpHandler   `yaml:"httpHandler"`
	Log         Log           `yaml:"log"`
	Redis       RedisDatabase `yaml:"redis"`
}

type SQLDatabase struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Name         string `yaml:"name"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Retry        int    `yaml:"retry"`
	RetryTimeout int    `yaml:"retryTimeout"`
}

type HttpHandler struct {
	Port    int `yaml:"port"`
	Workers int `yaml:"workers"`
}

type Log struct {
	Level string `yaml:"level"`
}

type RedisDatabase struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	DB           int    `yaml:"db"`
	Password     string `yaml:"password"`
	Retry        int    `yaml:"retry"`
	RetryTimeout int    `yaml:"retryTimeout"`
	TTL          int    `yaml:"TTL"`
}

func Init() (Config, error) {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	var err error
	defaultConfig = new(Config)
	if err = v.ReadInConfig(); err != nil {
		log.Errorf("Read the config file fail: %s", err)

		return *defaultConfig, err
	}

	if err = v.Unmarshal(&defaultConfig); err != nil {
		log.Errorf("Unmarshal config failed: %s", err)
	}

	return *defaultConfig, err
}

func GetRedis() *RedisDatabase {
	return &defaultConfig.Redis
}
