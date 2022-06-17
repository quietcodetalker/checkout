package main

import "gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/util"

type Config struct {
	DB      util.DBConfig `mapstructure:"db" validate:"required"`
	Brokers []string      `mapstructure:"brokers" validate:"required"`
}
