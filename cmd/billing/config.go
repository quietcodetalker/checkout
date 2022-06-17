package main

import "gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/util"

type Config struct {
	DB struct {
		Master  util.DBConfig `mapstructure:"master" validate:"required"`
		Replica util.DBConfig `mapstructure:"replica" validate:"required"`
	} `mapstructure:"db" validate:"required"`
	Brokers       []string `mapstructure:"brokers" validate:"required"`
	RedisAddr     string   `mapstructure:"redisAddr" validate:"required"`
	RedisPassword string   `mapstructure:"redisPassword"`
}
