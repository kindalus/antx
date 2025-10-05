package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type PwdCommand struct{}

func (c *PwdCommand) GetName() string {
	return "pwd"
}

func (c *PwdCommand) GetDescription() string {
	return "Show current location as path"
}

func (c *PwdCommand) Execute(args []string) {
	breadcrumbs, err := client.GetBreadcrumbs(currentFolder)
	if err != nil {
		fmt.Println("Error getting breadcrumbs:", err)
		// Fallback to old behavior
		fmt.Printf("%s  %s\n", currentFolder, currentFolderName)
		return
	}

	// Build path from breadcrumbs
	var pathParts []string
	for _, node := range breadcrumbs {
		if node.Title != "" {
			pathParts = append(pathParts, node.Title)
		}
	}

	if len(pathParts) == 0 {
		fmt.Println("/")
	} else {
		fmt.Printf("/%s\n", strings.Join(pathParts, "/"))
	}
}

func (c *PwdCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&PwdCommand{})
}
