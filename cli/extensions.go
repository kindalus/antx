package cli

import (
	"fmt"
	"sort"

	"github.com/c-bata/go-prompt"
)

type ExtensionsCommand struct{}

func (c *ExtensionsCommand) GetName() string {
	return "extensions"
}

func (c *ExtensionsCommand) GetDescription() string {
	return "List all available extensions"
}

func (c *ExtensionsCommand) Execute(args []string) {
	extensions, err := client.ListExtensions()
	if err != nil {
		fmt.Println("Error listing extensions:", err)
		return
	}

	if len(extensions) == 0 {
		fmt.Println("No extensions available.")
		return
	}

	// Sort extensions alphabetically by name
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Name < extensions[j].Name
	})

	fmt.Printf("Available extensions (%d):\n", len(extensions))
	fmt.Println()

	for _, extension := range extensions {
		fmt.Printf("UUID: %s\n", extension.UUID)
		fmt.Printf("  Name: %s\n", extension.Name)
		if extension.Description != "" {
			fmt.Printf("  Description: %s\n", extension.Description)
		}
		if extension.RunAs != "" {
			fmt.Printf("  Run As: %s\n", extension.RunAs)
		}
		if len(extension.GroupsAllowed) > 0 {
			fmt.Printf("  Groups Allowed: %v\n", extension.GroupsAllowed)
		}
		if len(extension.Parameters) > 0 {
			fmt.Printf("  Parameters:\n")
			for _, param := range extension.Parameters {
				required := ""
				if param.Required {
					required = " (required)"
				}
				fmt.Printf("    - %s (%s)%s: %s\n", param.Name, param.Type, required, param.Description)
				if param.DefaultValue != nil {
					fmt.Printf("      Default: %v\n", param.DefaultValue)
				}
			}
		}
		if extension.ReturnType != "" {
			fmt.Printf("  Return Type: %s\n", extension.ReturnType)
		}
		if extension.ReturnDescription != "" {
			fmt.Printf("  Return Description: %s\n", extension.ReturnDescription)
		}
		fmt.Println()
	}
}

func (c *ExtensionsCommand) Suggest(d prompt.Document) []prompt.Suggest {
	// This command doesn't take arguments, so no suggestions needed
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&ExtensionsCommand{})
}
