package cli

import (
	"fmt"

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
