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
		folder = currentFolder
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
	for _, node := range nodes {
		title := node.Title

		if len(title) > 40 {
			title = fmt.Sprintf("%s...", title[:37])
		}

		fmt.Printf(" %-40s  %-12s  %5s  %s\n", title, node.UUID, node.HumanReadableSize(), node.Mimetype)
	}
}

func (c *LsCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&LsCommand{})
}
