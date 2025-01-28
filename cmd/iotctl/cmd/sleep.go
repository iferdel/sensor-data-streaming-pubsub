package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/iferdel/treanteyes/internal/validation"
	"github.com/spf13/cobra"
)

var sleepCmd = &cobra.Command{
	Use:   "sleep",
	Short: "Stop sensor from generating/sending more data",
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
			fmt.Println("sending sleep command to all sensors -- not yet implemented")
			return
		}

		url := fmt.Sprintf("%s/sensors/%s/sleep", API_URL, sensorSerialNumber)
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
	rootCmd.AddCommand(sleepCmd)
	sleepCmd.Flags().StringP("sensor", "s", "", "sensorid")
	sleepCmd.Flags().BoolP("all", "a", false, "sleep all sensors")
}
