package main

import (
	"os"
)

func main() {
	if err := Execute(); err != nil {
		printStderr("%v\n", err)
		os.Exit(1)
	}
}
