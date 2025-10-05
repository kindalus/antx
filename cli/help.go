package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

type HelpCommand struct{}

func (c *HelpCommand) GetName() string {
	return "help"
}

func (c *HelpCommand) GetDescription() string {
	return "Show this help message"
}

func (c *HelpCommand) Execute(args []string) {
	fmt.Println("Available commands:")
	for _, cmd := range commands {
		fmt.Printf("  %-20s - %s\n", cmd.GetName(), cmd.GetDescription())
	}
}

func (c *HelpCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&HelpCommand{})
}
