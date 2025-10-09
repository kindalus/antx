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
		fmt.Println("Description:")
		fmt.Println("  Create a duplicate of a node in the same location.")
		fmt.Println("  The duplicate will have the same parent as the original.")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  uuid  UUID of the node to duplicate")
		fmt.Println()
		fmt.Println("Special aliases:")
		fmt.Println("  .   Current node")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  duplicate abc123-def456-ghi789")
		fmt.Println("  duplicate .  # Duplicate current node")
		return
	}

	nodeUUID := args[0]

	// Validate source node exists and get its info
	sourceNode, err := client.GetNode(nodeUUID)
	if err != nil {
		fmt.Printf("Error: Cannot access node '%s': %v\n", nodeUUID, err)
		return
	}

	// Perform the duplicate operation
	duplicatedNode, err := client.DuplicateNode(nodeUUID)
	if err != nil {
		fmt.Printf("Error: Failed to duplicate node: %v\n", err)
		return
	}

	// Success message
	fmt.Printf("Node duplicated successfully\n")
	fmt.Printf("  Original: %s (%s)\n", sourceNode.Title, nodeUUID)
	fmt.Printf("  New:      %s (%s)\n", duplicatedNode.Title, duplicatedNode.UUID)
}

func (c *DuplicateCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
}

func init() {
	RegisterCommand(&DuplicateCommand{})
}
