package main

import (
	_ "embed"
	"os"
	"strings"

	"github.com/iferdel/sensor-data-streaming-server/cmd/iotctl/cmd"
)

const url = "http://localhost:8080/api/v1/"

//go:embed version.txt
var version string

func main() {
	err := cmd.Execute(strings.Trim(version, "\n"))
	if err != nil {
		os.Exit(1)
	}
}
