package main

import (
	"os"

	"github.com/jana-mind/bubbler/cmd/bubbler/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
