package main

import (
	"os"

	"github.com/thedataflows/confedit/cmd"
	log "github.com/thedataflows/go-lib-log"
)

var version = "dev"

func main() {
	log.SetGlobalLogger(log.GlobalLoggerBuilder().WithoutBuffering().Build())
	defer log.Close()
	err := cmd.Run(version, os.Args[1:])
	if err != nil {
		log.Errorf("main", err, "Command failed")
		os.Exit(1)
	}
}
