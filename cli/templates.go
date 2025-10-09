package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
)

type TemplatesCommand struct{}

func (c *TemplatesCommand) GetName() string {
	return "templates"
}

func (c *TemplatesCommand) GetDescription() string {
	return "List all available templates or download a specific template"
}

func (c *TemplatesCommand) Execute(args []string) {
	if len(args) == 0 {
		// List all templates
		templates, err := client.ListTemplates()
		if err != nil {
			fmt.Println("Error listing templates:", err)
			return
		}

		if len(templates) == 0 {
			fmt.Println("No templates available.")
			return
		}

		fmt.Println("Available templates:")
		fmt.Println()
		for _, template := range templates {
			fmt.Printf("UUID: %s\n", template.UUID)
			fmt.Printf("  Mimetype: %s\n", template.Mimetype)
			fmt.Printf("  Size: %d bytes\n", template.Size)
			fmt.Println()
		}
		return
	}

	// Download specific template
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

func (c *TemplatesCommand) Suggest(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")

	if len(args) <= 2 {
		// First argument - suggest template UUIDs
		word := d.GetWordBeforeCursor()

		// Get cached templates or fetch them
		templates, err := client.ListTemplates()
		if err != nil {
			return []prompt.Suggest{}
		}

		var suggests []prompt.Suggest
		for _, template := range templates {
			if strings.HasPrefix(strings.ToLower(template.UUID), strings.ToLower(word)) {
				suggests = append(suggests, prompt.Suggest{
					Text:        template.UUID,
					Description: fmt.Sprintf("Mimetype: %s, Size: %d bytes", template.Mimetype, template.Size),
				})
			}
		}
		return suggests
	}

	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&TemplatesCommand{})
}
