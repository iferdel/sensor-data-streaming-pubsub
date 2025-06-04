package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	api "github.com/iferdel/sensor-data-streaming-server/cmd/iotctl/client"
	"github.com/iferdel/sensor-data-streaming-server/cmd/iotctl/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const API_URL = "http://localhost:8080/api/v1"

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "iotctl",
	Short: "CLI Tool for Managing IoT Sensors via REST API",
	Long: `This CLI tool allows you to manage resources (sensors) 
It allows the use of keywords to alter the behaviour of the available sensors in the cluster. 
Every command has some flags such as sensor id or parameters related to the command itself.`,
}

// Execute executes the root command.
func Execute(currentVersion string) error {
	rootCmd.Version = currentVersion
	info := version.FetchUpdateInfo(currentVersion)
	defer info.PromptUpdateIfAvailable()
	ctx := version.WithContext(context.Background(), &info)
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.iot.yaml")
}

func readViperConfig(paths []string) error {
	for _, path := range paths {
		_, err := os.Stat(path)
		if err != nil {
			continue
		}
		viper.SetConfigFile(path)
	}
	return viper.ReadInConfig()
}

func initConfig() {
	viper.SetDefault("api_url", "http://localhost:8080/api/v1")
	viper.SetDefault("access_token", "")
	viper.SetDefault("refresh_token", "")
	viper.SetDefault("last_refresh", 0)

	if cfgFile != "" {
		// use config file specified from the flag
		viper.SetConfigFile(cfgFile)
		err := viper.ReadInConfig()
		cobra.CheckErr(err)
	} else {
		// use default config file
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		defaultPath := path.Join(home, ".iot.yaml")
		configPaths := []string{}
		configPaths = append(configPaths, path.Join(home, ".config", "iot", "iot.yaml"))
		configPaths = append(configPaths, defaultPath)
		if err := readViperConfig(configPaths); err != nil {
			viper.SafeWriteConfigAs(defaultPath)
			viper.SetConfigFile(defaultPath)
			err = viper.ReadInConfig()
			cobra.CheckErr(err)
		}
	}

	viper.SetEnvPrefix("iot")
	viper.AutomaticEnv() // read in environment variables that match
}

// chain multiple commands together
func compose(commands ...func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		for _, command := range commands {
			command(cmd, args)
		}
	}
}

// this function, when called before a command handler, will tell the user that an update is required
func requireUpdate(cmd *cobra.Command, args []string) {
	info := version.FromContext(cmd.Context())

	if info == nil {
		fmt.Fprintln(os.Stderr, "Failed to fetch update info. Are you within the iot network?")
	}

	if info.FailtedToFetch != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch update info: %s\n", info.FailtedToFetch.Error())
	}

	if info.IsUpdateRequired {
		info.PromptUpdateIfAvailable()
		os.Exit(1)
	}
}

func requireAuth(cmd *cobra.Command, args []string) {
	promptLoginAndExitIf := func(condition bool) {
		if condition {
			fmt.Fprintln(os.Stderr, "You must be logged in to use that command.")
			fmt.Fprintln(os.Stderr, "Please run 'iotctl login' first.")
			os.Exit(1)
		}
	}

	access_token := viper.GetString("access_token")
	promptLoginAndExitIf(access_token == "")

	// refresh logic. It refreshes the token only if it is getting stale (55 minutes old)
	last_refresh := viper.GetInt64("last_refresh")
	if time.Now().Add(-time.Minute*55).Unix() <= last_refresh {
		return
	}

	creds, err := api.FetchAccessToken()
	promptLoginAndExitIf(err != nil)
	if creds.AccessToken == "" || creds.RefreshToken == "" {
		promptLoginAndExitIf(err != nil)
	}

	viper.Set("access_token", creds.AccessToken)
	viper.Set("refresh_token", creds.RefreshToken)
	viper.Set("last_refresh", time.Now().Unix())

	err = viper.WriteConfig()
	promptLoginAndExitIf(err != nil)
}
