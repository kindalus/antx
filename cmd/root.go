package cmd

import (
	"os"

	"github.com/kindalus/antx/cli"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "antx [server url]",
	Short: "A shell-like CLI for Antbox",
	Long:  `A shell-like CLI for Antbox, providing commands to interact with the Antbox API.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverURL := args[0]
		apiKey, _ := cmd.Flags().GetString("api-key")
		root, _ := cmd.Flags().GetString("root")
		jwt, _ := cmd.Flags().GetString("jwt")
		debug, _ := cmd.Flags().GetBool("verbose")
		cli.Start(serverURL, apiKey, root, jwt, debug)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("api-key", "", "API key for authentication")
	rootCmd.PersistentFlags().String("root", "", "Root password for authentication")
	rootCmd.PersistentFlags().String("jwt", "", "JWT token for authentication")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable debug mode")
}
