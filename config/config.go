package config

import (
	"errors"
	"github.com/kelseyhightower/envconfig"
	"fmt"
)

type Config struct {
	Binary          string
	Token           string
	AuthorizedUsers []int
	Timeout         int
}


func New(prefix string) (*Config, error) {
	var c Config
	err := envconfig.Process(prefix, &c)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error processing prefix %s, %s", prefix, err))
	}

	if len(c.Binary) == 0 {
		c.Binary = "python3"
	}
	if c.Timeout == 0 {
		c.Timeout = 10
	}
	return &c, nil
}
