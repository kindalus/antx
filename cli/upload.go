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
	return "Upload a file to a folder, feature, or aspect"
}

func (c *UploadCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: upload [-f|-a|-u <uuid>] <file-path>")
		fmt.Println("  -f: Upload as feature")
		fmt.Println("  -a: Upload as aspect")
		fmt.Println("  -u <uuid>: Update existing file")
		return
	}

	var filePath string
	var uploadType string = "file" // Default to regular file upload
	var updateUUID string

	// Parse flags
	argIndex := 0
	for argIndex < len(args) {
		switch args[argIndex] {
		case "-f":
			uploadType = "feature"
			argIndex++
		case "-a":
			uploadType = "aspect"
			argIndex++
		case "-u":
			if argIndex+1 >= len(args) {
				fmt.Println("Usage: upload -u <uuid> <file-path>")
				return
			}
			uploadType = "update"
			updateUUID = args[argIndex+1]
			argIndex += 2
		default:
			// Rest of args are file path
			filePath = strings.Join(args[argIndex:], " ")
			goto parseComplete
		}
	}
parseComplete:

	if filePath == "" {
		fmt.Println("Error: file path is required")
		return
	}

	if strings.HasPrefix(filePath, `"`) && strings.HasSuffix(filePath, `"`) {
		filePath = filePath[1 : len(filePath)-1]
	}

	switch uploadType {
	case "feature":
		feature, err := client.UploadFeature(filePath, nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Feature %s uploaded successfully with UUID %s\n", filePath, feature.UUID)

	case "aspect":
		aspect, err := client.UploadAspect(filePath, nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Aspect %s uploaded successfully with UUID %s\n", filePath, aspect.UUID)

	case "update":
		metadata := map[string]any{
			"uuid": updateUUID,
		}
		node, err := client.CreateFile(filePath, metadata)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("File %s updated successfully for node %s\n", filePath, node.UUID)

	default: // Regular file upload
		metadata := map[string]any{
			"parent": currentFolder,
		}
		node, err := client.CreateFile(filePath, metadata)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("File %s uploaded successfully to node %s\n", filePath, node.UUID)
	}
}

func (c *UploadCommand) Suggest(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")

	if len(args) == 2 {
		// First argument - suggest flags
		word := d.GetWordBeforeCursor()
		if strings.HasPrefix(word, "-") {
			return []prompt.Suggest{
				{Text: "-f", Description: "Upload as feature"},
				{Text: "-a", Description: "Upload as aspect"},
				{Text: "-u", Description: "Update existing file"},
			}
		}
		// No flag, suggest file path
		if word == "" {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				word = filepath.Join(homeDir, "Downloads")
			}
		}
		return getFileSystemSuggestions(word)
	}

	if len(args) > 2 {
		// Check if we have -u flag that needs a UUID
		for i, arg := range args[1:] {
			if arg == "-u" && i+2 == len(args)-1 {
				// Next argument should be UUID
				return getNodeSuggestions(d.GetWordBeforeCursor(), func(node antbox.Node) bool {
					return node.Mimetype != "application/vnd.antbox.folder"
				})
			}
		}

		// Otherwise suggest file path
		word := d.GetWordBeforeCursor()
		if word == "" {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				word = filepath.Join(homeDir, "Downloads")
			}
		}
		return getFileSystemSuggestions(word)
	}

	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&UploadCommand{})
}
