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
// - rag [options] [message]: RAG chat with optional location context
//   Options: -l (location context)
//   Interactive: If no message provided, enters interactive session (exit with 'exit' or Ctrl+D)
// - chat [options] <agent_uuid> [message]: Chat with specific agent (always interactive)
//   Options: -t <temperature>, -m <max_tokens>
//   Interactive: Always enters interactive session. Message is sent first if provided. (exit with 'exit' or Ctrl+D)
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
	client       antbox.Antbox
	currentNode  antbox.Node
	currentNodes []antbox.Node
	cliHistory   []string

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

	// Resolve aliases in arguments (except for cd command with .. argument)
	for i, arg := range args {
		// Special case: don't resolve ".." for cd command to preserve navigation behavior
		if commandName == "cd" && arg == ".." {
			continue
		}
		args[i] = resolveAlias(arg)
	}

	if cmd, ok := commands[commandName]; ok {
		cmd.Execute(args)
	} else {
		fmt.Println("Unknown command: " + commandName)
	}

	// Add command to history AFTER execution (so currentNode is updated)
	addCommandToHistory(in)

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

	// Initialize current node and load cached data at startup
	initializeCurrentNodeAndCacheData()

	// Restore CLI state from saved configuration
	if err := restoreFromConfig(); err != nil {
		fmt.Printf("Note: Could not restore previous session: %v\n", err)
	}

	// Show breadcrumbs on startup
	showStartupBreadcrumbs()

	// Initial ls
	if cmd, ok := commands["ls"]; ok {
		cmd.Execute([]string{})
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionTitle("Antbox CLI"),
		prompt.OptionLivePrefix(func() (string, bool) {
			folderName := getCurrentFolderName()
			return fmt.Sprintf("%s # ", folderName), true
		}),
		prompt.OptionCompletionWordSeparator(" "),
		prompt.OptionMaxSuggestion(10),
		// Integrate saved history for up/down arrow navigation
		prompt.OptionHistory(cliHistory),
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

// initializeCurrentNodeAndCacheData initializes current node and loads cached data at startup
func initializeCurrentNodeAndCacheData() {
	fmt.Print("Initializing... ")

	// Initialize current node (start at root - will be overridden by restoreFromConfig if saved state exists)
	currentNode = antbox.Node{
		UUID:     "--root--",
		Title:    "root",
		Mimetype: "application/vnd.antbox.folder",
	}

	// Initialize empty history
	cliHistory = []string{}

	// Load current folder contents
	if nodes, err := client.ListNodes("--root--"); err == nil {
		currentNodes = nodes
	}

	// Load cached data
	loadCachedData()

	fmt.Println("✓ Ready")
}

// showStartupBreadcrumbs displays the current location path on startup
func showStartupBreadcrumbs() {
	breadcrumbs, err := client.GetBreadcrumbs(currentNode.UUID)
	if err != nil {
		// Fallback to simple display
		fmt.Printf("Current location: %s\n\n", getCurrentFolderName())
		return
	}

	// Build path from breadcrumbs
	var pathParts []string
	for _, node := range breadcrumbs {
		if node.Title != "" {
			pathParts = append(pathParts, node.Title)
		}
	}

	if len(pathParts) == 0 {
		fmt.Printf("Current location: /\n\n")
	} else {
		fmt.Printf("Current location: /%s\n\n", strings.Join(pathParts, "/"))
	}
}

// loadCachedData loads aspects, actions, extensions, and agents
func loadCachedData() {
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

	// Report results quietly during initialization
	if len(failed) > 0 {
		fmt.Printf(" ✗ Failed: %s", strings.Join(failed, ", "))
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

// getCurrentFolderName returns the display name for the current folder
func getCurrentFolderName() string {
	if currentNode.UUID == "--root--" {
		return "root"
	}
	return currentNode.Title
}

// resolveAlias resolves special aliases to actual UUIDs
// . -> current node UUID
// .. -> parent node UUID
func resolveAlias(arg string) string {
	switch arg {
	case ".":
		return currentNode.UUID
	case "..":
		if currentNode.Parent == "" {
			return "--root--"
		}
		return currentNode.Parent
	default:
		return arg
	}
}

// GetCachedAspects returns the cached list of aspects
func GetCachedAspects() []antbox.Aspect {
	return cachedAspects
}
