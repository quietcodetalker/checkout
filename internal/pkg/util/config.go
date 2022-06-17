package util

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// DBConfig represents a common db configuration.
type DBConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Name     string `mapstructure:"name" validate:"required"`
	SSLMode  string `mapstructure:"sslmode" validate:"required"`
}

// LoadConfig loads yaml config and populates provided config struct.
func LoadConfig(path string, name string, v interface{}) error {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read err: %w", err)
	}

	err = viper.Unmarshal(&v)
	if err != nil {
		return fmt.Errorf("unmarshal err: %w", err)
	}

	validate := validator.New()
	err = validate.Struct(v)
	if err != nil {
		return fmt.Errorf("validation err: %w", err)
	}

	return nil
}
