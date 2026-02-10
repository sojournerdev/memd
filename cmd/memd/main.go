package main

import (
	"os"

	"github.com/sojournerdev/memd/internal/commands"
)

func main() {
	os.Exit(commands.Run(os.Args, os.Stdout, os.Stderr))
}
