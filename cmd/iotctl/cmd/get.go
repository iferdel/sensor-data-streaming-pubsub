package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iferdel/treanteyes/internal/storage"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve all registered sensors",
	Run: func(cmd *cobra.Command, args []string) {

		url := fmt.Sprintf("%s/sensors", API_URL)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("error making request: %w", err)
			return
		}
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		params := []storage.SensorRecord{}
		err = decoder.Decode(&params)
		if err != nil {
			fmt.Println(err)
			return
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fmt.Printf("received non-2xx response code: %d", resp.StatusCode)
			return
		}

		fmt.Println("Active sensors")
		for _, param := range params {
			fmt.Println("serial_number:", param.SerialNumber)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringP("sensor", "s", "", "sensorid")
}
