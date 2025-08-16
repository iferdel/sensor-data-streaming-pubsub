package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const logo string = `
 IIIII  OOOOO   OOOOO  OOOOO
   I      O    O     O   O  
   I      O    O     O   O  
   I      O    O     O   O  
 IIIII  OOOOO   OOOOO    O  
`

func logoRenderer() string {
	return logo
}

var loginCmd = &cobra.Command{
	Use:          "login",
	Aliases:      []string{"auth", "authenticate", "signin"},
	Short:        "Authenticate the CLI to IIoT API server",
	SilenceUsage: true,
	PreRun:       requireUpdate,
	RunE: func(cmd *cobra.Command, args []string) error {
		w, _, err := term.GetSize(0)
		if err != nil {
			w = 0
		}

		welcome := lipgloss.PlaceHorizontal(lipgloss.Width(logo), lipgloss.Center, "This is the IIoT CLI")

		if w >= lipgloss.Width(welcome) {
			fmt.Println("hal")
			fmt.Print(welcome, "\n\n")
		} else {
			fmt.Print("This is the IIoT CLI\n\n")
		}

		_ = viper.GetString("api_url") + "/cli/login"
		return nil

	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
