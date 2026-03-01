package main

import (
	"fmt"
	"os"

	ddcmd "github.com/metalagman/ddgo/cmd/ddsync/cmd"
)

func main() {
	if err := ddcmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
