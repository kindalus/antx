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
		fmt.Println("Description:")
		fmt.Println("  Copy a node to another location with an optional new title.")
		fmt.Println("  If no title is provided, generates 'Copy of <original_title>'.")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  source_uuid       UUID of the node to copy")
		fmt.Println("  destination_uuid  UUID of the destination folder")
		fmt.Println("  new_title         Optional new title for the copied node")
		fmt.Println()
		fmt.Println("Special aliases:")
		fmt.Println("  .   Current node (for source or destination)")
		fmt.Println("  ..  Parent node (for destination)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  cp abc123 def456")
		fmt.Println("  cp abc123 . \"Local Copy\"")
		fmt.Println("  cp . folder-uuid \"Copy of Current\"")
		fmt.Println("  cp doc-uuid .. \"Moved Up Copy\"")
		return
	}

	sourceUUID := args[0]
	destinationUUID := args[1]

	// Validate source node exists and get its info
	sourceNode, err := client.GetNode(sourceUUID)
	if err != nil {
		fmt.Printf("Error: Cannot access source node '%s': %v\n", sourceUUID, err)
		return
	}

	// Validate destination folder exists
	destNode, err := client.GetNode(destinationUUID)
	if err != nil {
		fmt.Printf("Error: Cannot access destination '%s': %v\n", destinationUUID, err)
		return
	}

	// Check if destination is a folder
	if destNode.Mimetype != "application/vnd.antbox.folder" && destNode.Mimetype != "application/vnd.antbox.smartfolder" {
		fmt.Printf("Error: Destination '%s' is not a folder (mimetype: %s)\n", destinationUUID, destNode.Mimetype)
		return
	}

	// Determine the title for the copied node
	var newTitle string
	if len(args) >= 3 {
		// Use provided title
		newTitle = strings.Join(args[2:], " ")
	} else {
		// Generate default title
		newTitle = "Copy of " + sourceNode.Title
	}

	// Perform the copy operation
	copiedNode, err := client.CopyNode(sourceUUID, destinationUUID, newTitle)
	if err != nil {
		fmt.Printf("Error: Failed to copy node: %v\n", err)
		return
	}

	// Success message
	fmt.Printf("Node copied successfully\n")
	fmt.Printf("  From: %s (%s)\n", sourceNode.Title, sourceUUID)
	fmt.Printf("  To:   %s (%s)\n", destNode.Title, destinationUUID)
	fmt.Printf("  New:  %s (%s)\n", copiedNode.Title, copiedNode.UUID)
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
