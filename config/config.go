package config

import (
	"github.com/spf13/viper"

	"url-shortner/log"
)

type Config struct {
	Database    SQLDatabase   `yaml:"database"`
	HttpHandler HttpHandler   `yaml:"httpHandler"`
	Log         Log           `yaml:"log"`
	Redis       RedisDatabase `yaml:"redisDatabase"`
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
	Port int `yaml:"port"`
}

type Log struct {
	Level string `yaml:"level"`
}

type RedisDatabase struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DB       int    `yaml:"db"`
	Password string `yaml:"password"`
}

func Init() (Config, error) {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	var err error
	cfg := new(Config)
	if err = v.ReadInConfig(); err != nil {
		log.Errorf("read the config file fail: %s", err)

		return *cfg, err
	}

	if err = v.Unmarshal(&cfg); err != nil {
		log.Errorf("Unmarshal config failed: %s", err)
	}

	return *cfg, err
}
