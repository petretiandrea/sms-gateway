package main

import (
	_ "embed"
	"os"

	"go.uber.org/zap"
)

//go:embed version.txt
var version string

func main() {
	log := zap.Must(zap.NewProduction()).Sugar()
	zap.ReplaceGlobals(log.Desugar())
	defer log.Sync()

	if err := newRootCommand(version, log).Execute(); err != nil {
		log.Errorw("command failed", "error", err)
		os.Exit(1)
	}
}
