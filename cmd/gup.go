package main

import (
	"fmt"
	"os"

	cli "github.com/herosql/get-auto-update/internal"
	log "github.com/herosql/get-auto-update/pkg/log"
)

func main() {
	cleanup, err := log.Init("info.log", "error.log")
	if err != nil {
		panic(err)
	}
	defer cleanup()

	err = cli.Update()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}
