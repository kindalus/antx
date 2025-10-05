package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/kindalus/antx/antbox"

	prompt "github.com/c-bata/go-prompt"
)

// Suggestion Window Behavior:
// - Tab: Navigate to next suggestion
// - Shift+Tab: Navigate to previous suggestion
// - Up/Down arrows: Navigate through suggestions
// - Enter: Select current suggestion and close window
// - Any other key: Select current suggestion and close window
// - Escape: Close window without selecting
//
// Auto-Hide Behavior:
// - Window disappears when you continue typing past a complete command
// - Window disappears when command is fully typed (exact match)
// - Window disappears after adding space following a complete argument
// - Window disappears on multiple consecutive spaces
// - Use Ctrl+Space to manually hide suggestions
//
// Available Commands:
// - pwd: Show current path using breadcrumbs
// - rag [options] <message>: RAG chat with optional filters and conversation context
//   Options: -l (location context), -c <conversation_id>, -f <field>=<value>
// - chat [options] <agent_uuid> <message>: Chat with specific agent
//   Options: -t <temperature>, -m <max_tokens>, -c <conversation_id>
// - answer [options] <agent_uuid> <question>: Ask question to specific agent
//   Options: -t <temperature>, -m <max_tokens>

var (
	client            antbox.Antbox
	currentFolder     = "--root--"
	currentFolderName = "root"
	currentNodes      []antbox.Node
)

func executor(in string) {
	in = strings.TrimSpace(in)
	parts := strings.Split(in, " ")
	commandName := parts[0]
	args := parts[1:]

	if cmd, ok := commands[commandName]; ok {
		cmd.Execute(args)
	} else {
		fmt.Println("Unknown command: " + commandName)
	}

	fmt.Println("")
}

func completer(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()

	// Hide suggestions if text is empty or just whitespace
	if strings.TrimSpace(text) == "" {
		return []prompt.Suggest{}
	}

	args := strings.Split(text, " ")
	commandName := args[0]

	// If we're typing the first word (command name)
	if len(args) == 1 {
		// Hide suggestions if command is followed by space (user moved on)
		if strings.HasSuffix(text, " ") {
			return []prompt.Suggest{}
		}

		// Check if this is an exact command match - if so, hide suggestions
		if _, exists := commands[commandName]; exists {
			return []prompt.Suggest{}
		}

		// Only show suggestions for partial matches
		var suggests []prompt.Suggest
		for name, cmd := range commands {
			if strings.HasPrefix(name, commandName) && name != commandName {
				suggests = append(suggests, prompt.Suggest{Text: name, Description: cmd.GetDescription()})
			}
		}
		return suggests
	}

	// For arguments: only show if command exists and we're actively typing
	if cmd, ok := commands[commandName]; ok {
		// Hide if multiple consecutive spaces (user finished typing)
		if strings.Contains(text, "  ") {
			return []prompt.Suggest{}
		}

		// Hide if last character is space and previous wasn't (user just added space)
		if len(text) > 1 && strings.HasSuffix(text, " ") && !strings.HasSuffix(text[:len(text)-1], " ") {
			// Get current word being typed
			lastWord := ""
			if len(args) > 1 {
				lastWord = args[len(args)-1]
			}

			// Only hide if the last word looks complete (no partial typing)
			if lastWord == "" || len(lastWord) > 2 {
				return []prompt.Suggest{}
			}
		}

		return cmd.Suggest(d)
	}

	return []prompt.Suggest{}
}

func Start(serverURL, apiKey, root, jwt string, debug bool) {
	client = antbox.NewClient(serverURL, apiKey, root, jwt, debug)
	if root != "" {
		if err := client.Login(); err != nil {
			fmt.Println("Login failed:", err)
			os.Exit(1)
		}
	}

	// Initial ls
	if cmd, ok := commands["ls"]; ok {
		cmd.Execute([]string{})
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionTitle("Antbox CLI"),
		prompt.OptionLivePrefix(func() (string, bool) {
			return fmt.Sprintf("%s # ", currentFolderName), true
		}),
		prompt.OptionCompletionWordSeparator(" "),
		prompt.OptionMaxSuggestion(10),
		// Add custom key bindings for better completion control
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.Escape,
			Fn: func(buf *prompt.Buffer) {
				// Escape key will close completion window without selecting
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlSpace,
			Fn: func(buf *prompt.Buffer) {
				// Ctrl+Space to force hide suggestions
			},
		}),
	)
	p.Run()
}
