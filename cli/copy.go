package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type CopyCommand struct{}

func (c *CopyCommand) GetName() string {
	return "cp"
}

func (c *CopyCommand) GetDescription() string {
	return "Copy a node to another location"
}

func (c *CopyCommand) Execute(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: cp <source_uuid> <destination_uuid> [new_title]")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  source_uuid: UUID of the node to copy")
		fmt.Println("  destination_uuid: UUID of the destination folder")
		fmt.Println("  new_title: Optional new title for the copied node")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  cp abc123 def456")
		fmt.Println("  cp abc123 def456 \"Copy of Document\"")
		return
	}

	sourceUUID := args[0]
	destinationUUID := args[1]

	// Determine the title for the copied node
	var newTitle string
	if len(args) >= 3 {
		// Use provided title
		newTitle = strings.Join(args[2:], " ")
	} else {
		// Get original node to generate a default title
		sourceNode, err := client.GetNode(sourceUUID)
		if err != nil {
			fmt.Println("Error getting source node:", err)
			return
		}
		newTitle = "Copy of " + sourceNode.Title
	}

	// Perform the copy operation
	copiedNode, err := client.CopyNode(sourceUUID, destinationUUID, newTitle)
	if err != nil {
		fmt.Println("Error copying node:", err)
		return
	}

	fmt.Printf("Node copied successfully:\n")
	fmt.Printf("  Source: %s\n", sourceUUID)
	fmt.Printf("  Destination: %s\n", destinationUUID)
	fmt.Printf("  New UUID: %s\n", copiedNode.UUID)
	fmt.Printf("  Title: %s\n", copiedNode.Title)
}

func (c *CopyCommand) Suggest(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	args := strings.Fields(text)

	if len(args) == 0 {
		return []prompt.Suggest{}
	}

	// Count actual arguments (excluding the command name)
	argCount := len(args) - 1
	if !strings.HasSuffix(text, " ") && len(args) > 1 {
		argCount = len(args) - 2 // We're still typing the current argument
	}

	switch argCount {
	case 0:
		// Suggesting source UUID - any node
		return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
	case 1:
		// Suggesting destination UUID - only folders
		return getNodeSuggestions(d.GetWordBeforeCursor(), folderFilter)
	default:
		// No suggestions for title
		return []prompt.Suggest{}
	}
}

func init() {
	RegisterCommand(&CopyCommand{})
}
