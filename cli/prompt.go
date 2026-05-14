package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kindalus/antx/antbox"

	prompt "github.com/c-bata/go-prompt"
)

// Suggestion Window Behavior:
// - Tab: Navigate to next suggestion OR trigger suggestions if none shown
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
// - Tab can be used to manually trigger suggestions at any time
//
// Available Commands:
// - pwd: Show current path using breadcrumbs
// - /[agent_uuid] [message]: Chat with a specific agent (always interactive)
// - /<agent_uuid> [message]: Alternate unbracketed chat shortcut
// - @[agent_uuid] <question>: Ask a specific agent for a single answer
// - @<agent_uuid> <question>: Alternate unbracketed answer shortcut
// - cp <source_uuid> <destination_uuid> [new_title]: Copy a node to another location
// - clone <uuid>: Clone a node in the same location
// - reload: Reload cached data from server (agents)
// - status: Show cached data statistics

var (
	client       antbox.Antbox
	currentNode  antbox.Node
	currentNodes []antbox.Node
	cliHistory   []string

	// Cached data loaded at startup
	cachedAgents []antbox.Agent
)

func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}

	if executeAgentShortcut(in) {
		addCommandToHistory(in)
		fmt.Println("")
		return
	}

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

func executeAgentShortcut(in string) bool {
	mode, agentUUID, message, matched := parseAgentShortcut(in)
	if !matched {
		return false
	}

	switch mode {
	case "chat":
		if agentUUID == "" {
			fmt.Println("Usage: /<agent_uuid> [message]")
			fmt.Println("   or: /[agent_uuid] [message]")
			return true
		}
		(&agentChatRunner{}).startInteractiveSession(agentUUID, message, nil, nil)
	case "answer":
		if agentUUID == "" || message == "" {
			fmt.Println("Usage: @<agent_uuid> <question>")
			fmt.Println("   or: @[agent_uuid] <question>")
			return true
		}
		(&agentAnswerRunner{}).askAgent(agentUUID, message, nil, nil)
	}

	return true
}

func parseAgentShortcut(input string) (mode, agentUUID, message string, matched bool) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", "", false
	}

	marker := input[0]
	switch marker {
	case '/':
		mode = "chat"
	case '@':
		mode = "answer"
	default:
		return "", "", "", false
	}

	remainder := strings.TrimSpace(input[1:])
	if remainder == "" {
		return mode, "", "", true
	}

	token, rest, hasRest := strings.Cut(remainder, " ")
	token = strings.TrimSpace(token)
	agentUUID = strings.TrimSuffix(strings.TrimPrefix(token, "["), "]")
	if hasRest {
		message = strings.TrimSpace(rest)
	}

	return mode, agentUUID, message, true
}

func getAgentShortcutSuggestions(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	if text == "" {
		return []prompt.Suggest{}
	}

	marker := text[0]
	if marker != '/' && marker != '@' {
		return []prompt.Suggest{}
	}

	parts := strings.Fields(text)
	if len(parts) > 1 || strings.HasSuffix(text, " ") {
		return []prompt.Suggest{}
	}

	currentWord := strings.TrimPrefix(d.GetWordBeforeCursor(), string(marker))
	currentWord = strings.TrimSuffix(strings.TrimPrefix(currentWord, "["), "]")

	var suggests []prompt.Suggest
	for _, agent := range GetCachedAgents() {
		if strings.HasPrefix(strings.ToLower(agent.UUID), strings.ToLower(currentWord)) ||
			strings.HasPrefix(strings.ToLower(agent.DisplayName()), strings.ToLower(currentWord)) {
			description := "Chat with " + agent.DisplayName()
			if marker == '@' {
				description = "Ask " + agent.DisplayName()
			}
			suggests = append(suggests, prompt.Suggest{
				Text:        fmt.Sprintf("%c[%s]", marker, agent.UUID),
				Description: description,
			})
		}
	}
	return suggests
}

func completer(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()

	if strings.HasPrefix(text, "/") || strings.HasPrefix(text, "@") {
		return getAgentShortcutSuggestions(d)
	}

	// Check if Tab was the last keystroke - if so, force show suggestions
	if d.LastKeyStroke() == prompt.Tab || d.LastKeyStroke() == prompt.ControlI {
		// For Tab, we want to show suggestions even when they would normally be hidden
		args := strings.Split(text, " ")
		commandName := strings.TrimSpace(args[0])

		// Case 1: Empty text or just whitespace - show all commands
		if strings.TrimSpace(text) == "" {
			var suggests []prompt.Suggest
			for name, cmd := range commands {
				suggests = append(suggests, prompt.Suggest{Text: name, Description: cmd.GetDescription()})
			}
			return suggests
		}

		// Case 2: Single word (command name) - show matching commands
		if len(args) == 1 && !strings.HasSuffix(text, " ") {
			var suggests []prompt.Suggest
			for name, cmd := range commands {
				// Show all commands that match the prefix (including exact matches)
				if strings.HasPrefix(name, commandName) {
					suggests = append(suggests, prompt.Suggest{Text: name, Description: cmd.GetDescription()})
				}
			}
			return suggests
		}

		// Case 3: Command followed by space or arguments - show command suggestions
		if cmd, ok := commands[commandName]; ok {
			return cmd.Suggest(d)
		}

		// Case 4: Invalid command - show all commands as fallback
		var suggests []prompt.Suggest
		for name, cmd := range commands {
			suggests = append(suggests, prompt.Suggest{Text: name, Description: cmd.GetDescription()})
		}
		return suggests
	}

	// Normal suggestion logic - hide suggestions if text is empty or just whitespace
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

func newClientForStartOptions(ctx context.Context, options StartOptions, output io.Writer) (antbox.Antbox, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}

	serverURL := options.ServerURL
	jwt := options.JWT
	if options.AuthLightray {
		apiURL, err := lightrayAPIURL(options.ServerURL)
		if err != nil {
			return nil, err
		}

		token, err := authenticateWithLightray(ctx, options.ServerURL, options.effectiveLightrayClientID(), output)
		if err != nil {
			return nil, fmt.Errorf("Lightray authentication failed: %w", err)
		}

		serverURL = apiURL
		jwt = token
	}

	configuredClient := antbox.NewClient(serverURL, options.APIKey, options.Root, jwt, options.Debug)
	if options.Root != "" {
		if err := configuredClient.Login(); err != nil {
			return nil, fmt.Errorf("login failed: %w", err)
		}
	}

	return configuredClient, nil
}

func Start(serverURL, apiKey, root, jwt string, debug bool) {
	StartWithOptions(StartOptions{
		ServerURL: serverURL,
		APIKey:    apiKey,
		Root:      root,
		JWT:       jwt,
		Debug:     debug,
	})
}

func StartWithOptions(options StartOptions) {
	configuredClient, err := newClientForStartOptions(context.Background(), options, os.Stdout)
	if err != nil {
		fmt.Println("Startup failed:", err)
		os.Exit(1)
	}
	client = configuredClient

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

// loadCachedData loads agents
func loadCachedData() {
	if agents, err := client.ListAgents(); err == nil {
		cachedAgents = agents
	} else {
		fmt.Printf(" ✗ Failed: agents")
	}
}

// reloadCachedData reloads all cached data from the server
func reloadCachedData() error {
	fmt.Print("Reloading resources from server... ")

	agents, err := client.ListAgents()
	if err != nil {
		fmt.Printf("done with errors\n")
		fmt.Printf("  Failed to load: agents\n")
		fmt.Printf("    agents: %v\n", err)
		return fmt.Errorf("failed to load agents")
	}

	cachedAgents = agents
	fmt.Printf("done (%d agents)\n", len(agents))
	return nil
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
