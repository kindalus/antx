package cli

import (
	"fmt"

	prompt "github.com/c-bata/go-prompt"
	"github.com/kindalus/antx/antbox"
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
		// Go to root
		currentNode = antbox.Node{
			UUID:     "--root--",
			Title:    "root",
			Mimetype: "application/vnd.antbox.folder",
		}
	} else {
		var targetUUID string

		// Handle special case: ".." means navigate to parent (original behavior)
		if args[0] == ".." {
			if currentNode.UUID == "--root--" {
				return // Already at root
			}
			if currentNode.Parent == "" || currentNode.Parent == "--root--" {
				// Parent is root
				currentNode = antbox.Node{
					UUID:     "--root--",
					Title:    "root",
					Mimetype: "application/vnd.antbox.folder",
				}
				// List contents of new current folder
				if cmd, ok := commands["ls"]; ok {
					cmd.Execute([]string{})
				}
				return
			} else {
				targetUUID = currentNode.Parent
			}
		} else {
			// For any other argument (including resolved aliases), treat as UUID to navigate to
			targetUUID = args[0]
		}

		// Get target node and navigate to it
		node, err := client.GetNode(targetUUID)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		currentNode = *node
	}

	// List contents of new current folder
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
