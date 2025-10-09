package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/kindalus/antx/antbox"
)

type LsCommand struct{}

func (c *LsCommand) GetName() string {
	return "ls"
}

func (c *LsCommand) GetDescription() string {
	return "List content of a folder"
}

func (c *LsCommand) Execute(args []string) {
	var folder string
	if len(args) > 0 {
		folder = args[0]
	} else {
		folder = currentNode.UUID
	}

	var nodes []antbox.Node
	var err error

	if folder != "--root--" {
		folderNode, err := client.GetNode(folder)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if folderNode.Mimetype == "application/vnd.antbox.smartfolder" {
			nodes, err = client.EvaluateNode(folder)
		} else {
			nodes, err = client.ListNodes(folder)
		}
	} else {
		nodes, err = client.ListNodes(folder)
	}

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	currentNodes = nodes

	// Sort nodes: directories first, then files, both alphabetically by title
	sortedNodes := sortNodesForListing(nodes)

	// Print header
	fmt.Printf(" %-12s  %4s  %-12s  %-30s  %s\n", "UUID", "SIZE", "MODIFIED", "MIMETYPE", "TITLE")
	fmt.Printf(" %-12s  %4s  %-12s  %-30s  %s\n", "----", "----", "--------", "--------", "-----")

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

func (c *LsCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&LsCommand{})
}
