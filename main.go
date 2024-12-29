package main

import (
	"os"

	"github.com/vectier/otter/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
