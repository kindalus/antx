package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

type PwdCommand struct{}

func (c *PwdCommand) GetName() string {
	return "pwd"
}

func (c *PwdCommand) GetDescription() string {
	return "Show current location"
}

func (c *PwdCommand) Execute(args []string) {
	fmt.Printf("%s  %s\n", currentFolder, currentFolderName)
}

func (c *PwdCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&PwdCommand{})
}
