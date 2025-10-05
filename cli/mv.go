package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type MvCommand struct{}

func (c *MvCommand) GetName() string {
	return "mv"
}

func (c *MvCommand) GetDescription() string {
	return "Move a node to another location"
}

func (c *MvCommand) Execute(args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: mv <uuid> <destination-uuid>")
		return
	}

	err := client.MoveNode(args[0], args[1])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Node %s moved to %s successfully\n", args[0], args[1])
}

func (c *MvCommand) Suggest(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")
	if len(args) == 2 {
		return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
	} else if len(args) == 3 {
		return getNodeSuggestions(d.GetWordBeforeCursor(), folderFilter)
	}
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&MvCommand{})
}
