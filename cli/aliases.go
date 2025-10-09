package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

type AliasesCommand struct{}

func (c *AliasesCommand) GetName() string {
	return "aliases"
}

func (c *AliasesCommand) GetDescription() string {
	return "Show current alias values"
}

func (c *AliasesCommand) Execute(args []string) {
	if len(args) > 0 {
		fmt.Println("Usage: aliases")
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println("  Display the current values of special aliases used in navigation.")
		fmt.Println("  These aliases can be used in any command that accepts UUIDs.")
		fmt.Println()
		fmt.Println("Available aliases:")
		fmt.Println("  .   Current node UUID")
		fmt.Println("  ..  Parent node UUID")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  stat .           # Show info about current node")
		fmt.Println("  run action-uuid .. # Run action on parent node")
		fmt.Println("  cd .             # Stay in current folder")
		return
	}

	fmt.Println("Current Alias Values:")
	fmt.Println("=====================")
	fmt.Printf("  .  (current) = %s", currentNode.UUID)
	if currentNode.UUID == "--root--" {
		fmt.Printf(" (root)")
	} else {
		fmt.Printf(" (%s)", currentNode.Title)
	}
	fmt.Println()

	parentUUID := currentNode.Parent
	if parentUUID == "" {
		parentUUID = "--root--"
	}
	fmt.Printf("  .. (parent)  = %s", parentUUID)

	if parentUUID == "--root--" {
		fmt.Printf(" (root)")
	} else {
		// Try to get parent node title for display
		if parentNode, err := client.GetNode(parentUUID); err == nil {
			fmt.Printf(" (%s)", parentNode.Title)
		}
	}
	fmt.Println()

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  These aliases are automatically resolved in all commands.")
	fmt.Println("  Example: 'stat .' shows info about the current node.")
	fmt.Println("  Example: 'cd ..' navigates to the parent folder.")
}

func (c *AliasesCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&AliasesCommand{})
}
