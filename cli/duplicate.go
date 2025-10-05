package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

type DuplicateCommand struct{}

func (c *DuplicateCommand) GetName() string {
	return "duplicate"
}

func (c *DuplicateCommand) GetDescription() string {
	return "Duplicate a node in the same location"
}

func (c *DuplicateCommand) Execute(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: duplicate <uuid>")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  uuid: UUID of the node to duplicate")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  duplicate abc123-def456-ghi789")
		return
	}

	nodeUUID := args[0]

	// Perform the duplicate operation
	duplicatedNode, err := client.DuplicateNode(nodeUUID)
	if err != nil {
		fmt.Println("Error duplicating node:", err)
		return
	}

	fmt.Printf("Node duplicated successfully:\n")
	fmt.Printf("  Original UUID: %s\n", nodeUUID)
	fmt.Printf("  New UUID: %s\n", duplicatedNode.UUID)
	fmt.Printf("  Title: %s\n", duplicatedNode.Title)
}

func (c *DuplicateCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
}

func init() {
	RegisterCommand(&DuplicateCommand{})
}
