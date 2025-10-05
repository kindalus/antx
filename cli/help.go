package cli

import (
	"fmt"
	"sort"
	"strings"

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
	// If a specific command is requested, show detailed help
	if len(args) > 0 {
		cmdName := args[0]
		if cmd, exists := commands[cmdName]; exists {
			fmt.Printf("Command: %s\n", cmd.GetName())
			fmt.Printf("Description: %s\n", cmd.GetDescription())
			fmt.Println()

			// Execute the command with no args to show its usage
			fmt.Println("Usage:")
			cmd.Execute([]string{})
		} else {
			fmt.Printf("Unknown command: %s\n", cmdName)
			fmt.Println()
			fmt.Println("Available commands:")

			// Show just a simple alphabetical list for unknown commands
			var cmdNames []string
			for name := range commands {
				cmdNames = append(cmdNames, name)
			}
			sort.Strings(cmdNames)

			for _, name := range cmdNames {
				fmt.Printf("  %s\n", name)
			}
		}
		return
	}

	fmt.Println("Antbox CLI - Available Commands")
	fmt.Println("===============================")
	fmt.Println()

	// Define command categories
	categories := map[string][]string{
		"Navigation & Browsing": {"cd", "ls", "pwd", "find", "stat"},
		"File Operations":       {"cp", "duplicate", "mv", "rename", "rm", "upload", "download"},
		"Folder Management":     {"mkdir", "mksmart"},
		"Actions & Extensions":  {"run", "call"},
		"AI & Agents":           {"chat", "answer", "rag"},
		"Templates":             {"template"},
		"System Management":     {"reload", "status", "help", "exit"},
	}

	// Print commands by category
	for _, category := range []string{
		"Navigation & Browsing",
		"File Operations",
		"Folder Management",
		"Actions & Extensions",
		"AI & Agents",
		"Templates",
		"System Management",
	} {
		fmt.Printf("%s:\n", category)

		// Sort commands within category
		cmdNames := categories[category]
		sort.Strings(cmdNames)

		for _, name := range cmdNames {
			if cmd, exists := commands[name]; exists {
				fmt.Printf("  %-12s - %s\n", cmd.GetName(), cmd.GetDescription())
			}
		}
		fmt.Println()
	}

	// Show any uncategorized commands
	var uncategorized []string
	categorizedCommands := make(map[string]bool)

	// Mark all categorized commands
	for _, cmdList := range categories {
		for _, name := range cmdList {
			categorizedCommands[name] = true
		}
	}

	// Find uncategorized commands
	for name := range commands {
		if !categorizedCommands[name] {
			uncategorized = append(uncategorized, name)
		}
	}

	// Display uncategorized commands if any exist
	if len(uncategorized) > 0 {
		fmt.Println("Other Commands:")
		sort.Strings(uncategorized)
		for _, name := range uncategorized {
			cmd := commands[name]
			fmt.Printf("  %-12s - %s\n", cmd.GetName(), cmd.GetDescription())
		}
		fmt.Println()
	}

	fmt.Printf("Type 'help <command>' for detailed usage information. (%d commands total)\n", len(commands))
	fmt.Println("Use Tab completion for command and argument suggestions.")
}

func (c *HelpCommand) Suggest(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	args := strings.Fields(text)

	if len(args) == 0 {
		return []prompt.Suggest{}
	}

	// If we're typing the first argument after "help", suggest command names
	if len(args) == 2 || (len(args) == 1 && strings.HasSuffix(text, " ")) {
		var suggests []prompt.Suggest
		currentWord := d.GetWordBeforeCursor()

		// Get all command names and sort them
		var cmdNames []string
		for name := range commands {
			cmdNames = append(cmdNames, name)
		}
		sort.Strings(cmdNames)

		// Filter commands based on what's being typed
		for _, name := range cmdNames {
			if strings.HasPrefix(strings.ToLower(name), strings.ToLower(currentWord)) {
				cmd := commands[name]
				suggests = append(suggests, prompt.Suggest{
					Text:        name,
					Description: cmd.GetDescription(),
				})
			}
		}
		return suggests
	}

	return []prompt.Suggest{}
}

func init() {
	RegisterCommand(&HelpCommand{})
}
