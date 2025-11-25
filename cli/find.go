package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type FindCommand struct{}

func (c *FindCommand) GetName() string {
	return "find"
}

func (c *FindCommand) GetDescription() string {
	return "Find nodes using filter criteria"
}

func (c *FindCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: find <criteria>")
		fmt.Println("  Simple: find some text")
		fmt.Println("  Complex: find title == Document,owner ~= admin,size > 1000")
		return
	}

	searchText := strings.Join(args, " ")

	result, err := client.FindNodes(searchText, 20, 1)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if len(result.Nodes) == 0 {
		fmt.Println("No nodes found matching the criteria")
		return
	}

	fmt.Printf("Found %d nodes:\n", len(result.Nodes))
	fmt.Printf(" %-12s  %4s  %-12s  %-30s  %s\n", "UUID", "SIZE", "MODIFIED", "MIMETYPE", "TITLE")
	fmt.Printf(" %-12s  %4s  %-12s  %-30s  %s\n", "----", "----", "--------", "--------", "-----")

	// Sort nodes: directories first, then files, both alphabetically by title
	sortedNodes := sortNodesForListing(result.Nodes)

	for _, node := range sortedNodes {
		// Format UUID (first 12 characters)
		uuid := node.UUID
		if len(uuid) > 12 {
			uuid = uuid[:12]
		}

		// Format size with padding
		size := node.HumanReadableSize()

		// Format modified date
		modifiedAt := formatModifiedDate(node.ModifiedAt)

		// Format mimetype (max 30 characters with ellipsis)
		mimetype := node.Mimetype
		if len(mimetype) > 30 {
			mimetype = fmt.Sprintf("%s...", mimetype[:27])
		}

		// Title is free form (no truncation)
		title := node.Title

		fmt.Printf(" %-12s  %4s  %-12s  %-30s  %s\n", uuid, size, modifiedAt, mimetype, title)
	}
}

func (c *FindCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&FindCommand{})
}
