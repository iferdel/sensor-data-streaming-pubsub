package main

import (
	_ "embed"
	"os"
	"strings"

	"github.com/iferdel/sensor-data-streaming-server/cmd/iotctl/cmd"
)

//go:embed version.txt
var version string

func main() {
	err := cmd.Execute(strings.Trim(version, "\n"))
	if err != nil {
		os.Exit(1)
	}
}
