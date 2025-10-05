package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

type RmCommand struct{}

func (c *RmCommand) GetName() string {
	return "rm"
}

func (c *RmCommand) GetDescription() string {
	return "Remove a node"
}

func (c *RmCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: rm <uuid>")
		return
	}

	err := client.RemoveNode(args[0])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Node %s removed successfully\n", args[0])
}

func (c *RmCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
}

func init() {
	RegisterCommand(&RmCommand{})
}
