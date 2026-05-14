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
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL := args[0]
		apiKey, _ := cmd.Flags().GetString("api-key")
		root, _ := cmd.Flags().GetString("root")
		jwt, _ := cmd.Flags().GetString("jwt")
		debug, _ := cmd.Flags().GetBool("verbose")
		lightray, _ := cmd.Flags().GetBool("lightray")
		lightrayClientID, _ := cmd.Flags().GetString("lightray-client-id")

		options := cli.StartOptions{
			ServerURL:        serverURL,
			APIKey:           apiKey,
			Root:             root,
			JWT:              jwt,
			Debug:            debug,
			Lightray:         lightray,
			LightrayClientID: lightrayClientID,
		}
		if err := options.Validate(); err != nil {
			return err
		}

		cli.StartWithOptions(options)
		return nil
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
	rootCmd.PersistentFlags().Bool("lightray", false, "Authenticate through a Lightray browser/device flow and use <server url>/api")
	rootCmd.PersistentFlags().String("lightray-client-id", cli.DefaultLightrayClientID(), "Lightray OAuth client ID for device authentication")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable debug mode")
}
