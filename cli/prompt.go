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
// - run <action_uuid> <node_uuid> [param=value...]: Run an action on a node with optional parameters
// - call <extension_uuid> [param=value...]: Run an extension with optional parameters
// - template <uuid>: Download a template to Downloads folder
// - cp <source_uuid> <destination_uuid> [new_title]: Copy a node to another location
// - duplicate <uuid>: Duplicate a node in the same location
// - reload: Reload cached data from server (aspects, actions, extensions, agents)
// - status: Show cached data statistics

var (
	client            antbox.Antbox
	currentFolder     = "--root--"
	currentFolderName = "root"
	currentNodes      []antbox.Node

	// Cached data loaded at startup
	cachedAspects    []antbox.Aspect
	cachedActions    []antbox.Feature
	cachedExtensions []antbox.Feature
	cachedAgents     []antbox.Agent
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

	// Load cached data at startup
	loadCachedData()

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

// loadCachedData loads aspects, actions, extensions, and agents at startup
func loadCachedData() {
	fmt.Print("Loading available resources... ")

	var loaded []string
	var failed []string

	// Load aspects
	if aspects, err := client.ListAspects(); err == nil {
		cachedAspects = aspects
		loaded = append(loaded, fmt.Sprintf("%d aspects", len(aspects)))
	} else {
		failed = append(failed, "aspects")
	}

	// Load actions
	if actions, err := client.ListActions(); err == nil {
		cachedActions = actions
		loaded = append(loaded, fmt.Sprintf("%d actions", len(actions)))
	} else {
		failed = append(failed, "actions")
	}

	// Load extensions
	if extensions, err := client.ListExtensions(); err == nil {
		cachedExtensions = extensions
		loaded = append(loaded, fmt.Sprintf("%d extensions", len(extensions)))
	} else {
		failed = append(failed, "extensions")
	}

	// Load agents
	if agents, err := client.ListAgents(); err == nil {
		cachedAgents = agents
		loaded = append(loaded, fmt.Sprintf("%d agents", len(agents)))
	} else {
		failed = append(failed, "agents")
	}

	if len(failed) == 0 {
		fmt.Printf("done (%s)\n", strings.Join(loaded, ", "))
	} else {
		fmt.Printf("done with warnings\n")
		if len(loaded) > 0 {
			fmt.Printf("  Loaded: %s\n", strings.Join(loaded, ", "))
		}
		fmt.Printf("  Failed: %s\n", strings.Join(failed, ", "))
	}
}

// reloadCachedData reloads all cached data from the server
func reloadCachedData() error {
	fmt.Print("Reloading resources from server... ")

	var loaded []string
	var failed []string
	var errors []string

	// Reload aspects
	if aspects, err := client.ListAspects(); err == nil {
		cachedAspects = aspects
		loaded = append(loaded, fmt.Sprintf("%d aspects", len(aspects)))
	} else {
		failed = append(failed, "aspects")
		errors = append(errors, fmt.Sprintf("aspects: %v", err))
	}

	// Reload actions
	if actions, err := client.ListActions(); err == nil {
		cachedActions = actions
		loaded = append(loaded, fmt.Sprintf("%d actions", len(actions)))
	} else {
		failed = append(failed, "actions")
		errors = append(errors, fmt.Sprintf("actions: %v", err))
	}

	// Reload extensions
	if extensions, err := client.ListExtensions(); err == nil {
		cachedExtensions = extensions
		loaded = append(loaded, fmt.Sprintf("%d extensions", len(extensions)))
	} else {
		failed = append(failed, "extensions")
		errors = append(errors, fmt.Sprintf("extensions: %v", err))
	}

	// Reload agents
	if agents, err := client.ListAgents(); err == nil {
		cachedAgents = agents
		loaded = append(loaded, fmt.Sprintf("%d agents", len(agents)))
	} else {
		failed = append(failed, "agents")
		errors = append(errors, fmt.Sprintf("agents: %v", err))
	}

	if len(failed) == 0 {
		fmt.Printf("done (%s)\n", strings.Join(loaded, ", "))
		return nil
	} else {
		fmt.Printf("done with errors\n")
		if len(loaded) > 0 {
			fmt.Printf("  Successfully loaded: %s\n", strings.Join(loaded, ", "))
		}
		fmt.Printf("  Failed to load: %s\n", strings.Join(failed, ", "))
		for _, errMsg := range errors {
			fmt.Printf("    %s\n", errMsg)
		}
		return fmt.Errorf("%d resources failed to load", len(failed))
	}
}

// GetCachedActions returns the cached list of actions
func GetCachedActions() []antbox.Feature {
	return cachedActions
}

// GetCachedExtensions returns the cached list of extensions
func GetCachedExtensions() []antbox.Feature {
	return cachedExtensions
}

// GetCachedAgents returns the cached list of agents
func GetCachedAgents() []antbox.Agent {
	return cachedAgents
}

// GetCachedAspects returns the cached list of aspects
func GetCachedAspects() []antbox.Aspect {
	return cachedAspects
}
