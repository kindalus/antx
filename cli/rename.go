package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type RenameCommand struct{}

func (c *RenameCommand) GetName() string {
	return "rename"
}

func (c *RenameCommand) GetDescription() string {
	return "Change the name of a node"
}

func (c *RenameCommand) Execute(args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: rename <uuid> <new-name>")
		return
	}

	err := client.ChangeNodeName(args[0], args[1])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Node %s renamed to '%s' successfully\n", args[0], args[1])
}

func (c *RenameCommand) Suggest(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")
	if len(args) == 2 {
		return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
	}
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&RenameCommand{})
}
