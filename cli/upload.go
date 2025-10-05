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
	var updateUUID string
	var filePath string

	if len(args) > 1 && args[0] == "-u" {
		// Update mode
		updateUUID = args[1]
		if len(args) > 2 {
			filePath = strings.Join(args[2:], " ")
			if strings.HasPrefix(filePath, `"`) && strings.HasSuffix(filePath, `"`) {
				filePath = filePath[1 : len(filePath)-1]
			}
		} else {
			fmt.Println("Usage: upload -u <uuid> <file-path>")
			return
		}

		node, err := client.UpdateFile(updateUUID, filePath)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("File %s uploaded successfully to node %s\n", filePath, node.UUID)
	} else {
		// Create mode
		if len(args) > 0 {
			filePath = strings.Join(args, " ")
			if strings.HasPrefix(filePath, `"`) && strings.HasSuffix(filePath, `"`) {
				filePath = filePath[1 : len(filePath)-1]
			}
		} else {
			fmt.Println("Usage: upload <file-path>")
			return
		}

		node, err := client.CreateFile(filePath, currentFolder)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("File %s uploaded successfully as %s\n", filePath, node.Title)
	}
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
