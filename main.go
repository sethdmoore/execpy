package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/sethdmoore/execpy/config"
)

func main() {
	c := config.Get()
	spew.Dump(c)
}
