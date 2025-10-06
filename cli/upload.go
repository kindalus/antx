package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/kindalus/antx/antbox"
)

type UploadCommand struct{}

func (c *UploadCommand) GetName() string {
	return "upload"
}

func (c *UploadCommand) GetDescription() string {
	return "Upload a file to a folder"
}

func (c *UploadCommand) Execute(args []string) {
	var filePath string
	metadata := map[string]any{
		"parent": currentFolder,
	}

	if len(args) == 0 {
		fmt.Println("Usage: upload [-u <uuid>] <file-path>")
		return
	}

	if len(args) > 1 && args[0] == "-u" {
		if len(args) == 2 {
			fmt.Println("Usage: upload -u <uuid> <file-path>")
			return
		}
		// Update mode
		metadata["uuid"] = args[1]

		filePath = strings.Join(args[2:], " ")
	} else {
		filePath = strings.Join(args, " ")
	}

	if strings.HasPrefix(filePath, `"`) && strings.HasSuffix(filePath, `"`) {
		filePath = filePath[1 : len(filePath)-1]
	}

	node, err := client.CreateFile(filePath, metadata)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("File %s uploaded successfully to node %s\n", filePath, node.UUID)

}

func (c *UploadCommand) Suggest(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")

	if len(args) > 1 && args[1] == "-u" {
		// Update mode
		if len(args) == 3 {
			return getNodeSuggestions(d.GetWordBeforeCursor(), func(node antbox.Node) bool {
				return node.Mimetype != "application/vnd.antbox.folder"
			})
		}
		if len(args) == 4 {
			word := d.GetWordBeforeCursor()
			if word == "" {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					word = filepath.Join(homeDir, "Downloads")
				}
			}
			return getFileSystemSuggestions(word)
		}
	} else {
		// Create mode
		if len(args) == 2 {
			word := d.GetWordBeforeCursor()
			if word == "" {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					word = filepath.Join(homeDir, "Downloads")
				}
			}
			return getFileSystemSuggestions(word)
		}
	}

	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&UploadCommand{})
}
