package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c-bata/go-prompt"
)

type DownloadCommand struct{}

func (c *DownloadCommand) GetName() string {
	return "download"
}

func (c *DownloadCommand) GetDescription() string {
	return "Download a node to Downloads folder"
}

func (c *DownloadCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: download <uuid>")
		return
	}

	// Get node details to get the title for filename
	node, err := client.GetNode(args[0])
	if err != nil {
		fmt.Println("Error getting node details:", err)
		return
	}

	// Get user's Downloads directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	downloadPath := filepath.Join(homeDir, "Downloads", node.Title)

	err = client.DownloadNode(args[0], downloadPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Node '%s' downloaded to %s\n", node.Title, downloadPath)
}

func (c *DownloadCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
}

func init() {
	RegisterCommand(&DownloadCommand{})
}
