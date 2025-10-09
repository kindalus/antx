package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

type CloneCommand struct{}

func (c *CloneCommand) GetName() string {
	return "clone"
}

func (c *CloneCommand) GetDescription() string {
	return "Clone a node in the same location"
}

func (c *CloneCommand) Execute(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: clone <uuid>")
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println("  Create a clone of a node in the same location.")
		fmt.Println("  The clone will have the same parent as the original.")
		fmt.Println("  This is an alternative to the 'duplicate' command.")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  uuid  UUID of the node to clone")
		fmt.Println()
		fmt.Println("Special aliases:")
		fmt.Println("  .   Current node")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  clone abc123-def456-ghi789")
		fmt.Println("  clone .  # Clone current node")
		fmt.Println()
		fmt.Println("Note:")
		fmt.Println("  This command performs the same operation as 'duplicate'.")
		return
	}

	nodeUUID := args[0]

	// Validate source node exists and get its info
	sourceNode, err := client.GetNode(nodeUUID)
	if err != nil {
		fmt.Printf("Error: Cannot access node '%s': %v\n", nodeUUID, err)
		return
	}

	// Perform the clone operation (uses the same API as duplicate)
	clonedNode, err := client.DuplicateNode(nodeUUID)
	if err != nil {
		fmt.Printf("Error: Failed to clone node: %v\n", err)
		return
	}

	// Success message
	fmt.Printf("Node cloned successfully\n")
	fmt.Printf("  Original: %s (%s)\n", sourceNode.Title, nodeUUID)
	fmt.Printf("  Clone:    %s (%s)\n", clonedNode.Title, clonedNode.UUID)
}

func (c *CloneCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
}

func init() {
	RegisterCommand(&CloneCommand{})
}
