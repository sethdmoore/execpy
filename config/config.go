package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/sethdmoore/execpy/globals"
	"os"
)

type Config struct {
	Binary          string
	Token           string
	AuthorizedUsers []string
	Timeout         int
}

var c Config

func init() {
	err := envconfig.Process(globals.AppPrefix, &c)
	if err != nil {
		fmt.Printf("Error processing prefix %s, %s", globals.AppPrefix, err)
		os.Exit(2)
	}

	if len(c.Binary) == 0 {
		c.Binary = "python3"
	}
	if c.Timeout == 0 {
		c.Timeout = 10
	}
}

func Get() *Config {
	return &c
}
