package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"text/tabwriter"
	"time"
)

func sensorOutput(wg *sync.WaitGroup, sensorName string, interval time.Duration, seed int64) {
    defer wg.Done()
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
    var wg sync.WaitGroup
    
    wg.Add(2)
    go sensorOutput(&wg, "sensor1", 1 * time.Second, 99)
    go sensorOutput(&wg, "sensor2", 2 * time.Second, 99)
    wg.Wait()
}

