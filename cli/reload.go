package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

type ReloadCommand struct{}

func (c *ReloadCommand) GetName() string {
	return "reload"
}

func (c *ReloadCommand) GetDescription() string {
	return "Reload cached data from server"
}

func (c *ReloadCommand) Execute(args []string) {
	if len(args) > 0 {
		fmt.Println("Usage: reload")
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println("  Refresh the cached lists of aspects, actions, extensions, and agents")
		fmt.Println("  from the server. This is useful when new resources have been added")
		fmt.Println("  or modified on the server since the CLI was started.")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  reload")
		return
	}

	err := reloadCachedData()
	if err != nil {
		fmt.Printf("Reload completed with warnings: %v\n", err)
		return
	}

	fmt.Printf("Successfully reloaded all cached data:\n")
	fmt.Printf("  - %d aspects\n", len(GetCachedAspects()))
	fmt.Printf("  - %d actions\n", len(GetCachedActions()))
	fmt.Printf("  - %d extensions\n", len(GetCachedExtensions()))
	fmt.Printf("  - %d agents\n", len(GetCachedAgents()))
}

func (c *ReloadCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&ReloadCommand{})
}
