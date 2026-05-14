package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/kindalus/antx/antbox"
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
	if strings.HasSuffix(node.Mimetype, "folder") && node.Permissions != nil {

		fmt.Printf(template, "Permissions", "")
		if len(node.Permissions.Group) > 0 {
			fmt.Printf("  %-9s: %s\n", "Group", joinPermissions(node.Permissions.Group))
		}
		if len(node.Permissions.Authenticated) > 0 {
			fmt.Printf("  %-9s: %s\n", "Auth", joinPermissions(node.Permissions.Authenticated))
		}
		if len(node.Permissions.Anonymous) > 0 {
			fmt.Printf("  %-9s: %s\n", "Anonymous", joinPermissions(node.Permissions.Anonymous))
		}
	}

	fmt.Printf(template, "Size", node.HumanReadableSize())
	fmt.Printf(template, "Created at", node.CreatedAt)
	fmt.Printf(template, "Modified at", node.ModifiedAt)
}

func joinPermissions(permissions []antbox.Permission) string {
	values := make([]string, len(permissions))
	for i, permission := range permissions {
		values[i] = string(permission)
	}
	return strings.Join(values, ", ")
}

func (c *StatCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return getNodeSuggestions(d.GetWordBeforeCursor(), nil)
}

func init() {
	RegisterCommand(&StatCommand{})
}
