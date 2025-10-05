package cli

import (
	"fmt"

	prompt "github.com/c-bata/go-prompt"
)

type CdCommand struct{}

func (c *CdCommand) GetName() string {
	return "cd"
}

func (c *CdCommand) GetDescription() string {
	return "Change directory"
}

func (c *CdCommand) Execute(args []string) {
	if len(args) == 0 {
		currentFolder = "--root--"
		currentFolderName = "root"
	} else if args[0] == ".." {
		if currentFolder == "--root--" {
			return
		}
		node, err := client.GetNode(currentFolder)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		currentFolder = node.Parent
		if currentFolder == "--root--" {
			currentFolderName = "root"
		} else {
			parentNode, err := client.GetNode(currentFolder)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			currentFolderName = parentNode.Title
		}
	} else {
		currentFolder = args[0]
		node, err := client.GetNode(currentFolder)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		currentFolderName = node.Title
	}
	if cmd, ok := commands["ls"]; ok {
		cmd.Execute([]string{})
	}
}

func (c *CdCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return getNodeSuggestions(d.GetWordBeforeCursor(), folderFilter)
}

func init() {
	RegisterCommand(&CdCommand{})
}
