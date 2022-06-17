package main

import "gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/util"

type Config struct {
	WarehousesDB util.DBConfig `mapstructure:"warehousesDB" validate:"required"`
}
