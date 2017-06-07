package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/sethdmoore/execpy/globals"
	"os"
)

type Config struct {
	Token           string
	AuthorizedUsers []string
}

var c Config

func init() {
	err := envconfig.Process(globals.AppPrefix, &c)
	if err != nil {
		fmt.Printf("Error processing prefix %s, %s", globals.AppPrefix, err)
		os.Exit(2)
	}
}

func Get() *Config {
	return &c
}
