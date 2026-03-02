package main

import (
	"log"

	ddcmd "github.com/metalagman/ddgo/cmd/ddsync/cmd"
)

func main() {
	err := ddcmd.Execute("dev")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
