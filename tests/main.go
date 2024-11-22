package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

type SensorData struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

func main() {
	serverURL := "http://localhost:8080/endpoint"

	for {
		// Generate simulated data
		data := SensorData{
			Timestamp: time.Now(),
			Value:     generateSineWave(),
		}

		// Convert to JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			continue
		}

		// Send data to the server
		resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error sending data:", err)
			continue
		}

		fmt.Printf("Sent data: %+v, Response: %s\n", data, resp.Status)
		resp.Body.Close()

		// Sleep for 1 second before sending the next data point
		time.Sleep(1 * time.Second)
	}
}

// Generates a sine wave value for simulation
func generateSineWave() float64 {
	// Use the current time in seconds to simulate a wave
	seconds := float64(time.Now().UnixNano()) / 1e9
	return math.Sin(seconds) + rand.Float64()*0.1 // Add some noise
}

