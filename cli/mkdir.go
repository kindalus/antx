package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

type MkdirCommand struct{}

func (c *MkdirCommand) GetName() string {
	return "mkdir"
}

func (c *MkdirCommand) GetDescription() string {
	return "Create a directory"
}

func (c *MkdirCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: mkdir <name>")
		return
	}

	fn := strings.Join(args, " ")

	// Remove initial and final ' or " if present
	if (strings.HasPrefix(fn, "'") && strings.HasSuffix(fn, "'")) || (strings.HasPrefix(fn, "\"") && strings.HasSuffix(fn, "\"")) {
		fn = fn[1 : len(fn)-1]
	}

	_, err := client.CreateFolder(currentFolder, fn)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func (c *MkdirCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&MkdirCommand{})
}
