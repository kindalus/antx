package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/kindalus/antx/antbox"
)

type MksmartCommand struct{}

func (c *MksmartCommand) GetName() string {
	return "mksmart"
}

func (c *MksmartCommand) GetDescription() string {
	return "Create a smart folder"
}

func (c *MksmartCommand) Execute(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: mksmart <name> <field> <operator> [value]")
		fmt.Println("  Example: mksmart \"My Documents\" title match document")
		fmt.Println("  Example: mksmart \"Large Files\" size > 1000000")
		return
	}

	name := args[0]
	field := args[1]
	operator := args[2]

	var value string
	if len(args) > 3 {
		value = strings.Join(args[3:], " ")
	}

	// Remove initial and final ' or " if present
	if (strings.HasPrefix(name, "'") && strings.HasSuffix(name, "'")) || (strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"")) {
		name = name[1 : len(name)-1]
	}

	// Create the filter for the smart folder
	var filters antbox.NodeFilters1D
	if value != "" {
		convertedValue := convertValue(value)
		filters = antbox.NodeFilters1D{
			antbox.NodeFilter{field, antbox.FilterOperator(operator), convertedValue},
		}
	} else {
		filters = antbox.NodeFilters1D{
			antbox.NodeFilter{field, antbox.FilterOperator(operator), nil},
		}
	}

	_, err := client.CreateSmartFolder(currentNode.UUID, name, filters)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Smart folder '%s' created successfully\n", name)
}

func (c *MksmartCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&MksmartCommand{})
}
