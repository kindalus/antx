package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type StatCommand struct{}

func (c *StatCommand) GetName() string {
	return "stat"
}

func (c *StatCommand) GetDescription() string {
	return "Show node properties"
}

func (c *StatCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: stat <uuid>")
		return
	}

	node, err := client.GetNode(args[0])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	template := "%-11s: %s\n"

	fmt.Printf(template, "UUID", node.UUID)
	fmt.Printf(template, "Title", node.Title)
	fmt.Printf(template, "Mimetype", node.Mimetype)
	fmt.Printf(template, "Parent", node.Parent)
	fmt.Printf(template, "Owner", node.Owner)
	if node.Group != "" {
		fmt.Printf(template, "Group", node.Group)
	}

	// Show permissions if not a folder
	if strings.HasSuffix(node.Mimetype, "folder") {

		fmt.Printf(template, "Permissions", "")
		if len(node.Permissions.Group) > 0 {
			fmt.Printf("  %-9s: %s\n", "Group", strings.Join(node.Permissions.Group, ", "))
		}
		if len(node.Permissions.Authenticated) > 0 {
			fmt.Printf("  %-9s: %s\n", "Auth", strings.Join(node.Permissions.Authenticated, ", "))
		}
		if len(node.Permissions.Anonymous) > 0 {
			fmt.Printf("  %-9s: %s\n", "Anonymous", strings.Join(node.Permissions.Anonymous, ", "))
		}
	}

	fmt.Printf(template, "Size", node.HumanReadableSize())
	fmt.Printf(template, "Created at", node.CreatedAt)
	fmt.Printf(template, "Modified at", node.ModifiedAt)
}

func (c *StatCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
}

func init() {
	RegisterCommand(&StatCommand{})
}
