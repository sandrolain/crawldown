package main

import (
	"fmt"
	"os"
)

func printStdout(format string, args ...any) {
	if _, err := fmt.Fprintf(os.Stdout, format, args...); err != nil {
		return
	}
}

func printlnStdout() {
	if _, err := fmt.Fprintln(os.Stdout); err != nil {
		return
	}
}

func printStderr(format string, args ...any) {
	if _, err := fmt.Fprintf(os.Stderr, format, args...); err != nil {
		return
	}
}
