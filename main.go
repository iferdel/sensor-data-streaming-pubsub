package main

import (
	"fmt"
	"math/rand"
	"os"
	"text/tabwriter"
	"time"
)

func sensorOutput(sensorName string, interval time.Duration, seed int64) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    r := rand.New(rand.NewSource(seed))
    w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

    show := func(name string, accX, accY, accZ any) {
        fmt.Fprintf(w, "%s\t%v\t%v\t%v\n", name, accX, accY, accZ)
    }
    for range ticker.C {
        show(sensorName, r.Float64(), r.Float64(), r.Float64())
        w.Flush()
    }
}

func main() {
    sensorOutput("sensor1", 1 * time.Second, 99)
}

