package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/iferdel/sensor-data-streaming-server/internal/validation"
	"github.com/spf13/cobra"
)

var awakeCmd = &cobra.Command{
	Use:   "awake",
	Short: "Awake sensor from sleep and restart generating/sending data",
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

		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Printf("error retrieving all flag: %v", err)
			return
		}

		if all {
			fmt.Println("sending awake command to all sensors -- not yet implemented")
			return
		}

		url := fmt.Sprintf("%s/sensors/%s/awake", API_URL, sensorSerialNumber)
		req, err := http.NewRequest(http.MethodPut, url, nil)
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
	rootCmd.AddCommand(awakeCmd)
	awakeCmd.Flags().StringP("sensor", "s", "", "sensorid")
	awakeCmd.Flags().BoolP("all", "a", false, "awake all sensors")
}
