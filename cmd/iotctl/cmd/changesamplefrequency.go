package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/iferdel/sensor-data-streaming-server/internal/validation"
	"github.com/spf13/cobra"
)

var changeSampleFrequencyCmd = &cobra.Command{
	Use:   "changeSampleFrequency",
	Short: "Change the sample frequency [Hz] of a sensor",
	Run: func(cmd *cobra.Command, args []string) {

		sensorSerialNumber, err := cmd.Flags().GetString("sensor")
		if err != nil {
			log.Printf("error retrieving sensorid flag: %v", err)
			return
		}
		if sensorSerialNumber == "" {
			log.Printf("sensor serial number cannot be empty")
			return
		}
		if !validation.HasValidCharacters(sensorSerialNumber) {
			log.Printf("sensor serial number not valid")
			return
		}

		newSampleFrequency, err := cmd.Flags().GetFloat64("changeSampleFrequency")
		if err != nil {
			log.Printf("error retrieving new sample frequency flag: %v", err)
			return
		}
		if newSampleFrequency <= 0 {
			fmt.Println("sample frequency must be a float greater than 0")
			return
		}

		type parameters struct {
			NewSampleFrequency float64 `json:"new_sample_frequency"`
		}
		jsonData, err := json.Marshal(parameters{
			NewSampleFrequency: newSampleFrequency,
		})
		if err != nil {
			log.Fatalf("error marshaling JSON: %v", err)
			return
		}

		url := fmt.Sprintf("%s/sensors/%s/change-sample-frequency", API_URL, sensorSerialNumber)
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))

		req.Header.Set("Content-Type", "application/json")

		if err != nil {
			fmt.Println(err)
			return
		}

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("error making request: %w", err)
			return
		}
		defer res.Body.Close()

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			fmt.Printf("received non-2xx response code: %d", res.StatusCode)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(changeSampleFrequencyCmd)
	changeSampleFrequencyCmd.Flags().StringP("sensor", "s", "", "sensorid")
	changeSampleFrequencyCmd.Flags().Float64P("changeSampleFrequency", "f", 1.0, "sample frequency")
}
