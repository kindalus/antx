package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

type HistoryCommand struct{}

func (c *HistoryCommand) GetName() string {
	return "history"
}

func (c *HistoryCommand) GetDescription() string {
	return "Show command history"
}

func (c *HistoryCommand) Execute(args []string) {
	if len(args) > 0 {
		fmt.Println("Usage: history")
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println("  Display the last 20 commands from the CLI history.")
		fmt.Println("  History is automatically saved to ~/.antx and restored on startup.")
		fmt.Println()
		fmt.Println("Features:")
		fmt.Println("  - Shows up to 20 most recent commands")
		fmt.Println("  - Excludes help, status, aliases, and exit commands")
		fmt.Println("  - Persistent across CLI sessions")
		fmt.Println("  - Automatically saved after each command")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  history")
		return
	}

	if len(cliHistory) == 0 {
		fmt.Println("No command history available.")
		fmt.Println()
		fmt.Println("Commands will appear here as you use the CLI.")
		fmt.Println("History excludes: help, status, aliases, exit")
		return
	}

	fmt.Println("Command History:")
	fmt.Println("================")

	// Show history with line numbers
	for i, cmd := range cliHistory {
		fmt.Printf("%3d  %s\n", i+1, cmd)
	}

	fmt.Println()
	fmt.Printf("Showing %d of last %d commands\n", len(cliHistory), maxHistorySize)

	// Show config file location
	if configPath, exists, err := getConfigInfo(); err == nil {
		if exists {
			fmt.Printf("History saved in: %s\n", configPath)
		} else {
			fmt.Printf("History will be saved to: %s\n", configPath)
		}
	}
}

func (c *HistoryCommand) Suggest(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&HistoryCommand{})
}
