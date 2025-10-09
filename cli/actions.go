package cli

import (
	"fmt"
	"sort"

	"github.com/c-bata/go-prompt"
)

type ActionsCommand struct{}

func (c *ActionsCommand) GetName() string {
	return "actions"
}

func (c *ActionsCommand) GetDescription() string {
	return "List all available actions"
}

func (c *ActionsCommand) Execute(args []string) {
	actions, err := client.ListActions()
	if err != nil {
		fmt.Println("Error listing actions:", err)
		return
	}

	if len(actions) == 0 {
		fmt.Println("No actions available.")
		return
	}

	// Sort actions alphabetically by name
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Name < actions[j].Name
	})

	fmt.Printf("Available actions (%d):\n", len(actions))
	fmt.Println()

	for _, action := range actions {
		fmt.Printf("UUID: %s\n", action.UUID)
		fmt.Printf("  Name: %s\n", action.Name)
		if action.Description != "" {
			fmt.Printf("  Description: %s\n", action.Description)
		}
		fmt.Printf("  Run Manually: %v\n", action.RunManually)
		fmt.Printf("  Run on Creates: %v\n", action.RunOnCreates)
		fmt.Printf("  Run on Updates: %v\n", action.RunOnUpdates)
		if action.RunAs != "" {
			fmt.Printf("  Run As: %s\n", action.RunAs)
		}
		if action.Filters != nil {
			fmt.Printf("  Node Filtering: enabled\n")
		}
		if len(action.GroupsAllowed) > 0 {
			fmt.Printf("  Groups Allowed: %v\n", action.GroupsAllowed)
		}
		if len(action.Parameters) > 0 {
			fmt.Printf("  Parameters:\n")
			for _, param := range action.Parameters {
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
		fmt.Println()
	}
}

func (c *ActionsCommand) Suggest(d prompt.Document) []prompt.Suggest {
	// This command doesn't take arguments, so no suggestions needed
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&ActionsCommand{})
}
