package cli

import (
	"fmt"
	"os"

	"github.com/c-bata/go-prompt"
)

type ExitCommand struct{}

func (c *ExitCommand) GetName() string {
	return "exit"
}

func (c *ExitCommand) GetDescription() string {
	return "Exit the CLI"
}

func (c *ExitCommand) Execute(args []string) {
	fmt.Println("Bye!")
	os.Exit(0)
}

func (c *ExitCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&ExitCommand{})
}
