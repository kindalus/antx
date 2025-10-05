package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c-bata/go-prompt"
)

type TemplateCommand struct{}

func (c *TemplateCommand) GetName() string {
	return "template"
}

func (c *TemplateCommand) GetDescription() string {
	return "Download a template to Downloads folder"
}

func (c *TemplateCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: template <uuid>")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  uuid: UUID of the template to download")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  template abc123-def456-ghi789")
		return
	}

	templateUUID := args[0]

	// Get template data
	templateData, err := client.GetTemplate(templateUUID)
	if err != nil {
		fmt.Println("Error getting template:", err)
		return
	}

	// Get user's Downloads directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	// Create filename for the template (using UUID as base name)
	filename := fmt.Sprintf("template_%s.txt", templateUUID)
	downloadPath := filepath.Join(homeDir, "Downloads", filename)

	// Ensure Downloads directory exists
	downloadsDir := filepath.Join(homeDir, "Downloads")
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		fmt.Println("Error creating Downloads directory:", err)
		return
	}

	// Write template data to file
	err = os.WriteFile(downloadPath, templateData, 0644)
	if err != nil {
		fmt.Println("Error writing template file:", err)
		return
	}

	fmt.Printf("Template downloaded to %s\n", downloadPath)
}

func (c *TemplateCommand) Suggest(d prompt.Document) []prompt.Suggest {
	// For now, just provide a generic suggestion since we don't have template listing
	return []prompt.Suggest{
		{Text: "", Description: "Enter template UUID"},
	}
}

func init() {
	RegisterCommand(&TemplateCommand{})
}
